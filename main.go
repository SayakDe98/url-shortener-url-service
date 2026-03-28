package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"urlshortener/db"
	"urlshortener/migration"
	pb "urlshortener/proto"
	"urlshortener/server"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load("./.env")
		if err != nil {
			fmt.Println("No .env file present")
		}
	}
	// ---- MySQL ----
	// dsn := "user:password@tcp(host:3306)/db_name?parseTime=true&loc=Asia%2FKolkata"
	// db, err := sql.Open("mysql", dsn)
	// if err != nil {
	// 	log.Fatal("DB connection error:", err)
	// }
	// if err = db.Ping(); err != nil {
	// 	log.Fatal("DB unreachable:", err)
	// }
	db := db.InitDB()

	// ---- Redis ----
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis unreachable:", err)
	}

	// ---- Migrations ----
	migration.RunMigrations(db)

	// ---- gRPC server ----
	lis, err := net.Listen("tcp", ":"+os.Getenv("GRPC_PORT"))
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}
	// creates new grp server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor, // replaces the old gin middleware.RequestLogger()
		),
	)

	pb.RegisterURLShortenerServer(grpcServer, &server.URLShortenerServer{
		DB:  db,
		RDB: rdb,
		Ctx: ctx,
	})

	reflection.Register(grpcServer)

	log.Println("gRPC server listening on :" + os.Getenv("GRPC_PORT"))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Server error:", err)
	}
}
