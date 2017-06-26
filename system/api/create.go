package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ponzu-cms/ponzu/system/admin/upload"
	"github.com/ponzu-cms/ponzu/system/db"
	"github.com/ponzu-cms/ponzu/system/item"

	"github.com/gorilla/schema"
)

// Createable accepts or rejects external POST requests to endpoints such as:
// /api/content/create?type=Review
type Createable interface {
	// Create enables external clients to submit content of a specific type
	Create(http.ResponseWriter, *http.Request) error
}

// Trustable allows external content to be auto-approved, meaning content sent
// as an Createable will be stored in the public content bucket
type Trustable interface {
	AutoApprove(http.ResponseWriter, *http.Request) error
}

func createContentHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		log.Println("[Create] error:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	t := req.URL.Query().Get("type")
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	p, found := item.Types[t]
	if !found {
		log.Println("[Create] attempt to submit unknown type:", t, "from:", req.RemoteAddr)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	post := p()

	ext, ok := post.(Createable)
	if !ok {
		log.Println("[Create] rejected non-createable type:", t, "from:", req.RemoteAddr)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	ts := fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UnixNano()/int64(time.Millisecond))
	req.PostForm.Set("timestamp", ts)
	req.PostForm.Set("updated", ts)

	urlPaths, err := upload.StoreFiles(req)
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	for name, urlPath := range urlPaths {
		req.PostForm.Set(name, urlPath)
	}

	// check for any multi-value fields (ex. checkbox fields)
	// and correctly format for db storage. Essentially, we need
	// fieldX.0: value1, fieldX.1: value2 => fieldX: []string{value1, value2}
	fieldOrderValue := make(map[string]map[string][]string)
	for k, v := range req.PostForm {
		if strings.Contains(k, ".") {
			fo := strings.Split(k, ".")

			// put the order and the field value into map
			field := string(fo[0])
			order := string(fo[1])
			if len(fieldOrderValue[field]) == 0 {
				fieldOrderValue[field] = make(map[string][]string)
			}

			// orderValue is 0:[?type=Thing&id=1]
			orderValue := fieldOrderValue[field]
			orderValue[order] = v
			fieldOrderValue[field] = orderValue

			// discard the post form value with name.N
			req.PostForm.Del(k)
		}

	}

	// add/set the key & value to the post form in order
	for f, ov := range fieldOrderValue {
		for i := 0; i < len(ov); i++ {
			position := fmt.Sprintf("%d", i)
			fieldValue := ov[position]

			if req.PostForm.Get(f) == "" {
				for i, fv := range fieldValue {
					if i == 0 {
						req.PostForm.Set(f, fv)
					} else {
						req.PostForm.Add(f, fv)
					}
				}
			} else {
				for _, fv := range fieldValue {
					req.PostForm.Add(f, fv)
				}
			}
		}
	}

	hook, ok := post.(item.Hookable)
	if !ok {
		log.Println("[Create] error: Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// Let's be nice and make a proper item for the Hookable methods
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	dec.SetAliasTag("json")
	err = dec.Decode(post, req.PostForm)
	if err != nil {
		log.Println("Error decoding post form for edit handler:", t, err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	err = hook.BeforeAPICreate(res, req)
	if err != nil {
		log.Println("[Create] error calling BeforeAccept:", err)
		return
	}

	err = ext.Create(res, req)
	if err != nil {
		log.Println("[Create] error calling Accept:", err)
		return
	}

	err = hook.BeforeSave(res, req)
	if err != nil {
		log.Println("[Create] error calling BeforeSave:", err)
		return
	}

	// set specifier for db bucket in case content is/isn't Trustable
	var spec string

	// check if the content is Trustable should be auto-approved, if so the
	// content is immediately added to the public content API. If not, then it
	// is added to a "pending" list, only visible to Admins in the CMS and only
	// if the type implements editor.Mergable
	trusted, ok := post.(Trustable)
	if ok {
		err := trusted.AutoApprove(res, req)
		if err != nil {
			log.Println("[Create] error calling AutoApprove:", err)
			return
		}
	} else {
		spec = "__pending"
	}

	id, err := db.SetContent(t+spec+":-1", req.PostForm)
	if err != nil {
		log.Println("[Create] error calling SetContent:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// set the target in the context so user can get saved value from db in hook
	ctx := context.WithValue(req.Context(), "target", fmt.Sprintf("%s:%d", t, id))
	req = req.WithContext(ctx)

	err = hook.AfterSave(res, req)
	if err != nil {
		log.Println("[Create] error calling AfterSave:", err)
		return
	}

	err = hook.AfterAPICreate(res, req)
	if err != nil {
		log.Println("[Create] error calling AfterAccept:", err)
		return
	}

	// create JSON response to send data back to client
	var data map[string]interface{}
	if spec != "" {
		spec = strings.TrimPrefix(spec, "__")
		data = map[string]interface{}{
			"status": spec,
			"type":   t,
		}
	} else {
		spec = "public"
		data = map[string]interface{}{
			"id":     id,
			"status": spec,
			"type":   t,
		}
	}

	resp := map[string]interface{}{
		"data": []map[string]interface{}{
			data,
		},
	}

	j, err := json.Marshal(resp)
	if err != nil {
		log.Println("[Create] error marshalling response to JSON:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	_, err = res.Write(j)
	if err != nil {
		log.Println("[Create] error writing response:", err)
		return
	}

}
