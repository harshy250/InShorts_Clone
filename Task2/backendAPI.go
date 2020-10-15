package main

import(
	"fmt"
	"net/http"
	"time"
	"encoding/json"
	"strings"
	"sync"
	"io/ioutil"
) 

const (
	itemsPerPage = 2
)

type Article struct {
	Title 		string 	`json:"title"`
	ID 			string  `json:"id"`
	SubTitle 	string  `json:"subtitle"`
	Content 	string	`json:"content"`
	TimeStamp 	time.Time `json:"timestamp"`
}

type articleHandlers struct {
	sync.Mutex
	store map[string]Article
}

func (h *articleHandlers) articles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
		return
	case "POST":
		h.post(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}
}

// func (h *articleHandlers) pagination(w http.ResponseWriter, r *http.Request){
// 	start := (page-1) * itemsPerPage

// }

func (h *articleHandlers) get(w http.ResponseWriter, r *http.Request) {
	articles := make([]Article, len(h.store))

	h.Lock()
	i := 0
	for _, article := range h.store {
		articles[i] = article
		i++
	}
	h.Unlock()

	jsonBytes, err := json.Marshal(articles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *articleHandlers) post(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("need content-type 'application/json', but got '%s'", ct)))
		return
	}


	var article Article
	err = json.Unmarshal(bodyBytes, &article)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("need content-type 'application/json', but got '%s'", ct)))
		return
	}

	article.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	article.TimeStamp = time.Now()
	h.Lock()
	h.store[article.ID] = article
	defer h.Unlock()
}

func (h *articleHandlers) getArticle(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	article, ok := h.store[parts[2]]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jsonBytes, err := json.Marshal(article)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}	

func (h *articleHandlers) searchArticle(w http.ResponseWriter, r *http.Request){
	v := r.URL.Query().Get("q")

	var queries[]Article
	var titleArr[]string
	var subTitleArr[]string
	var contentArr[]string
	

	for _,item:= range h.store{
		// Title
		titleArr = strings.Split(item.Title, " ")
		for _,titem:= range titleArr{
			if(titem==v){
				queries=append(queries, item)
				break
			}
		}
		//Sub-Title
		subTitleArr = strings.Split(item.SubTitle, " ")
		for _,subtitem:= range subTitleArr{
			if(subtitem==v){
				queries=append(queries, item)
				break
			}
		}
		//Content
		contentArr = strings.Split(item.Content, " ")
		for _,contitem:= range contentArr{
			if(contitem==v){
				queries=append(queries, item)
				break
			}
		}
	}

	result, err:=json.Marshal(queries)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}


func newArticleHandlers() *articleHandlers {
	return &articleHandlers{
		store: map[string]Article{}, 
	}
}

func main() {
	articleHandlers := newArticleHandlers()
	http.HandleFunc("/articles", articleHandlers.articles)
	http.HandleFunc("/articles/", articleHandlers.getArticle)
	http.HandleFunc("/articles/search", articleHandlers.searchArticle)
	http.ListenAndServe(":8080", nil)
}