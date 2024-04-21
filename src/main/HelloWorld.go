package main

//import (
//	"log"
//	"net/http"
//)
//
//type server int
//
//func main() {
//	var s server
//	http.ListenAndServe("127.0.0.1:19326", &s)
//}
//
//func (h *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	log.Println(r.URL.Path)
//	_, err := w.Write([]byte("Hello World!"))
//	if err != nil {
//		return
//	}
//}
