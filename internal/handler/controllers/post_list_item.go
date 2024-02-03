package controllers

import (
	"github.com/fanky5g/ponzu/internal/application/storage"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"github.com/fanky5g/ponzu/internal/domain/services/management/editor"
	"log"
	"strings"
	"time"
)

// PostListItem is a helper to create the li containing a post.
// p is the asserted post as an Editable, t is the Type of the post.
// specifier is passed to append a name to a namespace like __pending
func PostListItem(e editor.Editable, typeName, status string) []byte {
	s, ok := e.(item.Sortable)
	if !ok {
		log.Println("Content type", typeName, "doesn't implement item.Sortable")
		post := `<li class="col s12">Error retrieving data. Your data type doesn't implement necessary interfaces. (item.Sortable)</li>`
		return []byte(post)
	}

	i, ok := e.(item.Identifiable)
	if !ok {
		log.Println("Content type", typeName, "doesn't implement item.Identifiable")
		post := `<li class="col s12">Error retrieving data. Your data type doesn't implement necessary interfaces. (item.Identifiable)</li>`
		return []byte(post)
	}

	// use sort to get other info to display in controllers UI post list
	tsTime := time.Unix(int64(s.Time()/1000), 0)
	upTime := time.Unix(int64(s.Touch()/1000), 0)
	updatedTime := upTime.Format("01/02/06 03:04 PM")
	publishTime := tsTime.Format("01/02/06")

	cid := i.ItemID()

	switch status {
	case "public", "":
		status = ""
	default:
		status = "__" + status
	}
	action := "/edit/delete"
	link := `<a href="/edit?type=` + typeName + `&status=` + strings.TrimPrefix(status, "__") + `&id=` + cid + `">` + i.String() + `</a>`
	if strings.HasPrefix(typeName, storage.UploadsEntityName) {
		link = `<a href="/edit/upload?id=` + cid + `">` + i.String() + `</a>`
		action = "/edit/upload/delete"
	}

	post := `
			<li class="col s12">
				` + link + `
				<span class="post-detail">Updated: ` + updatedTime + `</span>
				<span class="publish-date right">` + publishTime + `</span>

				<form enctype="multipart/form-data" class="quick-delete-post __ponzu right" action="` + action + `" method="post">
					<span>Delete</span>
					<input type="hidden" name="id" value="` + cid + `" />
					<input type="hidden" name="type" value="` + typeName + status + `" />
				</form>
			</li>`

	return []byte(post)
}
