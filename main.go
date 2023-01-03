package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"regexp"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"google.golang.org/grpc"

	protobuffer_comment "github.com/mxbikes/protobuf/comment"
	protobuffer_user "github.com/mxbikes/protobuf/user"
)

var (
	port              = getEnv("PORT", ":4105")
	URLServiceComment = getEnv("SERVICE_COMMENT_URL", "localhost:4089")
	URLServiceUser    = getEnv("SERVICE_USER_URL", "localhost:4090")
)

func main() {
	logger := &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.DebugLevel,
		Formatter: &prefixed.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			ForceFormatting: true,
		},
	}

	gwmux := runtime.NewServeMux()
	listener, err := net.Listen("tcp", port)
	if err != nil {
		logger.WithFields(logrus.Fields{"prefix": "REVERSE.PROXY"}).Fatalf("failed to listen: %v", err)
	}

	// Register ServiceComment
	err = protobuffer_comment.RegisterCommentServiceHandlerFromEndpoint(context.Background(), gwmux, URLServiceComment, []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		logger.WithFields(logrus.Fields{"prefix": "SERVICE.COMMENT"}).Fatal("unable to open a connection to service")
	}

	// Register ServiceUser
	err = protobuffer_user.RegisterUserServiceHandlerFromEndpoint(context.Background(), gwmux, URLServiceUser, []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		logger.WithFields(logrus.Fields{"prefix": "SERVICE.USER"}).Fatal("unable to open a connection to service")
	}

	gwServer := &http.Server{
		Addr:    port,
		Handler: cors(gwmux),
	}

	logger.WithFields(logrus.Fields{"prefix": "REVERSE.PROXY"}).Infof("is listening on Grpc PORT: {%v}", port)
	if err := gwServer.Serve(listener); err != nil {
		logger.WithFields(logrus.Fields{"prefix": "REVERSE.PROXY"}).Fatalf("failed to serve: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func allowedOrigin(origin string) bool {
	if viper.GetString("cors") == "*" {
		return true
	}
	if matched, _ := regexp.MatchString(viper.GetString("cors"), origin); matched {
		return true
	}
	return false
}

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowedOrigin(r.Header.Get("Origin")) {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, PUT")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType")
		}
		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
