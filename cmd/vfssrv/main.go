package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"vfs"
	"vfs/db"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-pg/pg"
	"github.com/semrush/zenrpc"
)

var (
	flAddr       = flag.String("addr", "localhost:9999", "listen address")
	flDir        = flag.String("dir", "testdata", "storage path")
	flNamespaces = flag.String("ns", "items,test", "namespaces, separated by comma")
	flWebPath    = flag.String("webpath", "/media/", "namespaces, separated by comma")
	flFileSize   = flag.Int64("maxsize", 32<<20, "max file size in bytes")
	flDbConn     = flag.String("conn", "postgresql://localhost:5432/vfs?sslmode=disable", "database connection dsn")
	flJWTKey     = flag.String("jwt.key", "QuiuNae9OhzoKohcee0h", "JWT key")
	flJWTHeader  = flag.String("jwt.header", "AuthorizationJWT", "JWT key")
	version      string
)

func main() {
	flag.Parse()

	v, err := vfs.New(vfs.Config{
		MaxFileSize:    *flFileSize,
		Path:           *flDir,
		WebPath:        *flWebPath,
		UploadFormName: "Filedata",
		Namespaces:     strings.Split(*flNamespaces, ","),
		Database:       nil,
	})

	if err != nil {
		log.Fatal(err)
	}

	repo := initRepo()

	rpc := zenrpc.NewServer(zenrpc.Options{ExposeSMD: true, AllowCORS: true})
	rpc.Use(zenrpc.Logger(log.New(os.Stderr, "", log.LstdFlags)))
	rpc.Register("", vfs.NewService(repo, v))
	http.Handle("/rpc", authMiddleware(rpc))

	http.HandleFunc("/auth-token", issueTokenHandler)
	http.Handle("/upload/hash", authMiddleware(http.HandlerFunc(v.HashUploadHandler)))
	http.Handle("/upload/file", authMiddleware(v.UploadHandler(repo)))
	http.Handle(*flWebPath, http.StripPrefix(*flWebPath, http.FileServer(http.Dir(*flDir))))

	log.Printf("starting vfssrv version=%v on %s", version, *flAddr)
	log.Fatal(http.ListenAndServe(*flAddr, nil))
}

// issueTokenHandler issues new jwt token for 1 hour. Subject can be set by id GET/POST param
func issueTokenHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "vfs",
		Subject:   id,
	})

	key := []byte(*flJWTKey)
	tokenString, err := token.SignedString(key)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		fmt.Fprint(w, tokenString)
		log.Printf("issued new token=%v for id=%v", tokenString, id)
	}
}

// authMiddleware checks JWT token if set in flag jwt.header.
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			isOK    = true
			errMsg  = ""
			errCode = http.StatusUnauthorized
		)

		defer func() {
			if isOK {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, errMsg, errCode)
			}
		}()

		if *flJWTHeader != "" {
			isOK = false
			tokenString := r.Header.Get(*flJWTHeader)
			token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(*flJWTKey), nil
			})

			if token == nil {
				errMsg = "missing token"
				return
			} else if err != nil {
				errMsg = err.Error()
			}

			// validate token
			if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
				if err = claims.Valid(); err != nil {
					errMsg, errCode = err.Error(), http.StatusForbidden
				} else {
					isOK = true
				}
			} else {
				errMsg = "bad token"
			}
		}
	})
}

// initRepo connects to postgres and inits vfs db repo.
func initRepo() db.VfsRepo {
	cfg, err := pg.ParseURL(*flDbConn)
	if err != nil {
		log.Fatal(err)
	}

	dbc := pg.Connect(cfg)
	d := db.New(dbc)
	v, err := d.Version()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(v)

	return db.NewVfsFileRepo(d)
}
