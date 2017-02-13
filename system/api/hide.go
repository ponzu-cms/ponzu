package api

import (
	"net/http"

	"github.com/ponzu-cms/ponzu/system/item"
)

func hide(it interface{}, res http.ResponseWriter, req *http.Request) bool {
	// check if should be hidden
	if h, ok := it.(item.Hideable); ok {
		err := h.Hide(res, req)
		if err == item.ErrAllowHiddenItem {
			return false
		}

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return true
		}

		res.WriteHeader(http.StatusNotFound)
		return true
	}

	return false
}
