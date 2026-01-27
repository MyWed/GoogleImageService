package main

import (
	"encoding/json"
	"fmt"
	"google.golang.org/appengine/v2"
	"google.golang.org/appengine/v2/blobstore"
	"google.golang.org/appengine/v2/image"
	"log"
	"net/http"
	"os"
	"strings"
)

type PostRequest struct {
	Files  []string `schema:"files,required"` // `json:"files"`
	Bucket string   `schema:"bucket"`         // `json:"bucket"`
}

type OutFile struct {
	Thumbnail string `json:"thumbnail"`
	Error     string `json:"error"`
}

func main() {
	http.HandleFunc("/", imageHandler)
	appengine.Main()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var post PostRequest
	// сначала нужно делать r.PostFormValue, там что-то инициализируется и тогда и цикл работает
	post.Bucket = r.PostFormValue("bucket")
	for key, value := range r.PostForm {
		if strings.HasPrefix(key, "files") {
			post.Files = append(post.Files, value[0])
		}
	}

	//fmt.Fprintf(w, "b%vb", post.Bucket) // bmywed-166514.appspot.comb
	//fmt.Fprintf(w, "b1%vb1", post.Bucket != "") // b1trueb1
	//fmt.Fprintf(w, "b2%vb2", post.Bucket == "") // b2falseb2

	// для запросов с JSON
	//if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
	//	http.Error(w, fmt.Sprintf("Could not decode body: %v", err), http.StatusBadRequest)
	//	return
	//}

	if len(post.Files) == 0 {
		fmt.Fprint(w, "hello world!")
	} else {
		myBucket := "mywed-166514.appspot.com"
		if post.Bucket != "" {
			myBucket = post.Bucket
		}
		//fmt.Fprintf(w, "m%vm", myBucket) // mmywed-166514.appspot.comm
		out := map[string]OutFile{}
		for _, file := range post.Files {
			filename := "/gs/" + myBucket + "/" + file
			//fmt.Fprintf(w, "f%vf", filename) // f/gs/mywed-166514.appspot.com/original/27/52a/70780155.jpgf
			blobKey, err := blobstore.BlobKeyForFile(c, filename)
			//fmt.Fprintf(w, "b%vb", blobKey) //
			if err == nil {
				url, urlErr := image.ServingURL(c, blobKey, &image.ServingURLOptions{Secure: true})
				if urlErr == nil {
					out[file] = OutFile{
						Thumbnail: url.String(),
						Error:     "",
					}
				} else {
					out[file] = OutFile{
						Thumbnail: "",
						Error:     urlErr.Error(),
					}
				}
			} else {
				//service bridge HTTP failed: Post "http://appengine.googleapis.internal:10001/rpc_http": dial tcp: lookup appengine.googleapis.internal on 169.254.169.254:53: no such host{}
				fmt.Fprint(w, err.Error())
			}
		}
		json.NewEncoder(w).Encode(out)
		w.Header().Set("Content-Type", "application/json")
	}
}
