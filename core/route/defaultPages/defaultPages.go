package defaultPages

import (
	"net/http"
)

func Page_404_get(w http.ResponseWriter) {
	page_404_get.Execute(w, nil)
}

func Page_404_post(w http.ResponseWriter) {
	page_404_post.Execute(w, nil)
}

func Page_405_get(w http.ResponseWriter) {
	page_405_get.Execute(w, nil)
}

func Page_405_post(w http.ResponseWriter) {
	page_405_post.Execute(w, nil)
}
