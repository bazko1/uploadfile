package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"uploadfile/myfile"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	dbFile := "files.db"
	createFileRouting(e, dbFile)
	e.Logger.Fatal(e.Start(":1323"))
}

func createFileRouting(e *echo.Echo, dbname string) {
	memcached, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(8192),
	)
	if err != nil {
		e.Logger.Panicf("failed to create cache: %v", err)
	}

	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(10*time.Minute),
		cache.ClientWithRefreshKey("opn"),
	)
	if err != nil {
		e.Logger.Panicf("failed to create cache: %v", err)
	}

	// This way we can test if caching is working fine
	// e.GET("/test", func(c echo.Context) error {
	// 	return c.Blob(http.StatusOK, "", []byte(time.Now().String()))
	// }, echo.WrapMiddleware(cacheClient.Middleware))

	controller, err := myfile.NewSQLiteController(dbname)
	if err != nil {
		e.Logger.Panicf("failed to initialize db conn: %v", err)
	}

	e.POST("/upload", upload(&controller))

	e.Use(middleware.BasicAuth(func(username string, password string, c echo.Context) (bool, error) {
		// Password hash should be validated from db or some thirdparty service.
		// Only printing for demo purpose.
		return true, nil
	}))

	t := &Template{
		templates: template.Must(template.ParseFiles("template/index.html")),
	}
	e.Renderer = t
	e.GET("/", index(&controller))

	g := e.Group("/api/v1/file")

	g.GET("", func(c echo.Context) error {
		files, err := controller.GetFilesMeta()
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, files)
	})

	// caching is added only for downloading full file contents as a blob
	g.GET("/download/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Logger().Errorf("failed to get file by id: %v", err)
			return c.NoContent(http.StatusNoContent)
		}
		file, err := controller.GetFileByID(id)
		if err != nil {
			c.Logger().Errorf("failed to get file by id: %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		return c.Blob(http.StatusOK, "application/octet-stream", file.Data)
	}, echo.WrapMiddleware(cacheClient.Middleware))

	g.POST("", func(c echo.Context) error {
		file := myfile.MyFile{}
		decoder := json.NewDecoder(c.Request().Body)
		if err := decoder.Decode(&file); err != nil {
			c.Logger().Errorf("failed to decode file: %v", err)
			return err
		}
		err := controller.AddFile(&file)
		if err != nil {
			c.Logger().Errorf("failed to add file: %v", err)
			return err
		}

		return c.NoContent(http.StatusOK)
	})
}

func upload(controller *myfile.FileController) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		ffile, err := c.FormFile("file")
		if err != nil {
			c.Logger().Errorf("failed to get form file: %v", err)
			return err
		}

		src, err := ffile.Open()
		if err != nil {
			c.Logger().Errorf("failed to open file: %v", err)
			return err
		}
		defer src.Close()

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(src)
		if err != nil {
			c.Logger().Errorf("failed to read file: %v", err)
			return err
		}

		f := myfile.MyFile{
			FileMeta: myfile.FileMeta{
				Name:       ffile.Filename,
				UploadedBy: name,
				Email:      email,
			},
			Data: buf.Bytes(),
		}

		err = controller.AddFile(&f)
		if err != nil {
			c.Logger().Errorf("controller failed to add file: %v", err)
			return err
		}

		return c.Redirect(http.StatusFound, "/")
	}
}

func index(controller *myfile.FileController) echo.HandlerFunc {
	return func(c echo.Context) error {
		files, err := controller.GetFilesMeta()
		if err != nil {
			c.Logger().Errorf("failed to get files meta: %v", err)
			return err
		}
		return c.Render(http.StatusOK, "index.html", files)
	}
}
