package main

import (
	"database/sql"
	"github.com/coopernurse/gorp"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
        "github.com/codegangsta/martini-contrib/binding"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
        "log"
        "time"
)


func main() {

    // initialize the DbMap
    dbmap := initDb()
    defer dbmap.Db.Close()

    // setup some of the database

    // delete any existing rows
    err := dbmap.TruncateTables()
    checkErr(err, "TruncateTables failed")

    // create two posts
    p1 := newPost("Post 1", "Lorem ipsum lorem ipsum")
    p2 := newPost("Post 2", "This is my second post")

    // insert rows
    err = dbmap.Insert(&p1, &p2)
    checkErr(err, "Insert failed")






    // lets start martini and the real code
    m := martini.Classic()

    m.Use(render.Renderer(render.Options{
    	Directory: "templates",
        Layout: "layout",
        Funcs: []template.FuncMap{
        	{
			"formatTime": func(args ...interface{}) string { 
				var s string
    				t1 := time.Unix(args[0].(int64), 0)
				s = t1.Format(time.Stamp)
    				return s
                        },
			"unescaped": func(args ...interface{}) template.HTML {
                                return template.HTML(args[0].(string))
                        },
                },
        },
    }))

    m.Get("/", func(r render.Render) {
        //fetch all rows
        var posts []Post
        _, err = dbmap.Select(&posts, "select * from posts order by post_id")
        checkErr(err, "Select failed")

        for x, p := range posts {
            log.Printf("    %d: %v\n", x, p)
        }

        newmap := map[string]interface{}{"metatitle": "this is my custom title", "posts": posts}

        r.HTML(200, "posts", newmap)
    })

    m.Get("/:id", func(args martini.Params, r render.Render) {
        var post Post

        err = dbmap.SelectOne(&post, "select * from posts where post_id=?", args["id"])
        
        //simple error check
        if err != nil {
          newmap := map[string]interface{}{"metatitle":"404 Error", "message":"This is not found"}
       	  r.HTML(404, "error", newmap)
        } else {
          newmap := map[string]interface{}{"metatitle": post.Title+" more custom", "post": post}
          r.HTML(200, "post", newmap)
        }
    })

    //shows how to create with binding params
    m.Post("/", binding.Bind(Post{}), func(post Post, r render.Render) {

        p1 := newPost(post.Title, post.Body)
        
        log.Println(p1)

        err = dbmap.Insert(&p1)
        checkErr(err, "Insert failed")
        
        newmap := map[string]interface{}{"metatitle": "created post", "post": p1}
        r.HTML(200, "post", newmap)
    })

    m.Run()

}


type Post struct {
    // db tag lets you specify the column name if it differs from the struct field
    Id      int64 `db:"post_id"`
    Created int64
    Title   string `form:"Title" binding:"required"`
    Body    string `form:"Body"`
}

func newPost(title, body string) Post {
    return Post{
        //Created: time.Now().UnixNano(),
        Created: time.Now().Unix(),
        Title:   title,
        Body:    body,
    }
}

func initDb() *gorp.DbMap {
    // connect to db using standard Go database/sql API
    // use whatever database/sql driver you wish
    
    //db, err := sql.Open("sqlite3", "/tmp/post_db.bin")
    db, err := sql.Open("mysql", "root:1meahstarr@unix(/var/run/mysqld/mysqld.sock)/sample")
    checkErr(err, "sql.Open failed")

    // construct a gorp DbMap
    // dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
    dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}

    // add a table, setting the table name to 'posts' and
    // specifying that the Id property is an auto incrementing PK
    dbmap.AddTableWithName(Post{}, "posts").SetKeys(true, "Id")

    // create the table. in a production system you'd generally
    // use a migration tool, or create the tables via scripts
    err = dbmap.CreateTablesIfNotExists()
    checkErr(err, "Create tables failed")

    return dbmap
}

func checkErr(err error, msg string) {
    if err != nil {
        log.Fatalln(msg, err)
    }
}
