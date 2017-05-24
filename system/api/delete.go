package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ponzu-cms/ponzu/system/db"
	"github.com/ponzu-cms/ponzu/system/item"
)

// Deleteable accepts or rejects update POST requests to endpoints such as:
// /api/content/delete?type=Review&id=1
type Deleteable interface {
	// Delete enables external clients to delete content of a specific type
	Delete(http.ResponseWriter, *http.Request) error
}

func deleteContentHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		log.Println("[Delete] error:", err)
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
		log.Println("[Delete] attempt to delete content of unknown type:", t, "from:", req.RemoteAddr)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	id := req.URL.Query().Get("id")
	if !db.IsValidID(id) {
		log.Println("[Delete] attempt to delete content with missing or invalid id from:", req.RemoteAddr)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	post := p()

	ext, ok := post.(Deleteable)
	if !ok {
		log.Println("[Delete] rejected non-deleteable type:", t, "from:", req.RemoteAddr)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	hook, ok := post.(item.Hookable)
	if !ok {
		log.Println("[Delete] error: Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	b, err := db.Content(t + ":" + id)
	if err != nil {
		log.Println("Error in db.Content ", t+":"+id, err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(b, post)
	if err != nil {
		log.Println("Error unmarshalling ", t, "=", id, err, " Hooks will be called on a zero-value.")
	}

	err = hook.BeforeAPIDelete(res, req)
	if err != nil {
		log.Println("[Delete] error calling BeforeAPIDelete:", err)
		if err == ErrNoAuth {
			// BeforeAPIDelete can check user.IsValid(req) for auth
			res.WriteHeader(http.StatusUnauthorized)
		}
		return
	}

	err = ext.Delete(res, req)
	if err != nil {
		log.Println("[Delete] error calling Delete:", err)
		if err == ErrNoAuth {
			// Delete can check user.IsValid(req) or other forms of validation for auth
			res.WriteHeader(http.StatusUnauthorized)
		}
		return
	}

	err = hook.BeforeDelete(res, req)
	if err != nil {
		log.Println("[Delete] error calling BeforeSave:", err)
		return
	}

	err = db.DeleteContent(t + ":" + id)
	if err != nil {
		log.Println("[Delete] error calling DeleteContent:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = hook.AfterDelete(res, req)
	if err != nil {
		log.Println("[Delete] error calling AfterDelete:", err)
		return
	}

	err = hook.AfterAPIDelete(res, req)
	if err != nil {
		log.Println("[Delete] error calling AfterAPIDelete:", err)
		return
	}

	// create JSON response to send data back to client
	var data = map[string]interface{}{
		"id":     id,
		"status": "deleted",
		"type":   t,
	}

	resp := map[string]interface{}{
		"data": []map[string]interface{}{
			data,
		},
	}

	j, err := json.Marshal(resp)
	if err != nil {
		log.Println("[Delete] error marshalling response to JSON:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	_, err = res.Write(j)
	if err != nil {
		log.Println("[Delete] error writing response:", err)
		return
	}

}
