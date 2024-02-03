package controllers

import (
	"bytes"
	"fmt"
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"github.com/fanky5g/ponzu/internal/domain/services/content"
	"github.com/fanky5g/ponzu/internal/domain/services/management/editor"
	"github.com/fanky5g/ponzu/internal/handler/controllers/views"
	"github.com/fanky5g/ponzu/internal/util"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Createable accepts or rejects external POST requests to endpoints such as:
// /api/content/create?type=Review
type Createable interface {
	// Create enables external clients to submit content of a specific type
	Create(http.ResponseWriter, *http.Request) error
}

func NewContentsHandler(configService config.Service, contentService content.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		appName, err := configService.GetAppName()
		if err != nil {
			log.Printf("Failed to get app name: %v\n", appName)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		q := req.URL.Query()
		t := q.Get("type")
		if t == "" {
			res.WriteHeader(http.StatusBadRequest)
			errView, err := views.Admin(util.Html("error_400"), appName)
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		order := strings.ToLower(q.Get("order"))
		if order != "asc" {
			order = "desc"
		}

		status := q.Get("status")

		if _, ok := item.Types[t]; !ok {
			res.WriteHeader(http.StatusBadRequest)
			errView, err := views.Admin(util.Html("error_400"), appName)
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		pt := item.Types[t]()
		if _, ok := pt.(editor.Editable); !ok {
			LogAndFail(res, fmt.Errorf("item %s does not implement editable interface", t), appName)
			return
		}

		var hasExt bool
		if _, ok := pt.(Createable); ok {
			hasExt = true
		}

		count, err := strconv.Atoi(q.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
		if err != nil {
			if q.Get("count") == "" {
				count = 10
			} else {
				LogAndFail(res, err, appName)
				return
			}
		}

		offset, err := strconv.Atoi(q.Get("offset")) // int: multiplier of count for pagination (0 default)
		if err != nil {
			if q.Get("offset") == "" {
				offset = 0
			} else {
				LogAndFail(res, err, appName)
				return
			}
		}

		opts := interfaces.QueryOptions{
			Count:  count,
			Offset: offset,
			Order:  order,
		}

		var specifier string
		if status == "public" || status == "" {
			specifier = "__sorted"
		} else if status == "pending" {
			specifier = "__pending"
		}

		b := &bytes.Buffer{}
		var total int
		var posts []interface{}

		html := `<div class="col s9 card">		
					<div class="card-content">
					<div class="row">
					<div class="col s8">
						<div class="row">
							<div class="card-title col s7">` + t + ` Items</div>
							<div class="col s5 input-field inline">
								<select class="browser-default __ponzu sort-order">
									<option value="DESC">New to Old</option>
									<option value="ASC">Old to New</option>
								</select>
								<label class="active">Sort:</label>
							</div>	
							<script>
								$(function() {
									var sort = $('select.__ponzu.sort-order');

									sort.on('change', function() {
										var path = window.location.pathname;
										var s = sort.val();
										var t = getParam('type');
										var status = getParam('status');

										if (status == "") {
											status = "public";
										}

										window.location.replace(path + '?type=' + t + '&order=' + s + '&status=' + status);
									});

									var order = getParam('order');
									if (order !== '') {
										sort.val(order);
									}
									
								});
							</script>
						</div>
					</div>
					<form class="col s4" action="/contents/search" method="get">
						<div class="input-field post-search inline">
							<label class="active">Search:</label>
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Within all ` + t + ` fields" class="search"/>
							<input type="hidden" name="type" value="` + t + `" />
							<input type="hidden" name="status" value="` + status + `" />
						</div>
                    </form>	
					</div>`
		if hasExt {
			if status == "" {
				q.Set("status", "public")
			}

			// always start from top of results when changing public/pending
			q.Del("count")
			q.Del("offset")

			q.Set("status", "public")
			publicURL := req.URL.Path + "?" + q.Encode()

			q.Set("status", "pending")
			pendingURL := req.URL.Path + "?" + q.Encode()

			switch status {
			case "public", "":
				// get __sorted posts of type t from the db
				total, posts, err = contentService.Query(t+specifier, opts)
				if err != nil {
					LogAndFail(res, err, appName)
					return
				}

				html += `<div class="row externalable">
					<span class="description">Status:</span> 
					<span class="active">Public</span>
					&nbsp;&vert;&nbsp;
					<a href="` + pendingURL + `">Pending</a>
				</div>`

				for _, entity := range posts {
					post := PostListItem(entity.(editor.Editable), t, status)
					_, err = b.Write(post)
					if err != nil {
						LogAndFail(res, err, appName)
						return
					}
				}

			case "pending":
				// get __pending posts of type t from the db
				total, posts, err = contentService.Query(t+"__pending", opts)
				if err != nil {
					LogAndFail(res, err, appName)
					return
				}

				html += `<div class="row externalable">
					<span class="description">Status:</span> 
					<a href="` + publicURL + `">Public</a>
					&nbsp;&vert;&nbsp;
					<span class="active">Pending</span>					
				</div>`

				for i := len(posts) - 1; i >= 0; i-- {
					post := PostListItem(posts[i].(editor.Editable), t, status)
					_, err = b.Write(post)
					if err != nil {
						LogAndFail(res, err, appName)
						return
					}
				}
			}

		} else {
			total, posts, err = contentService.Query(t+specifier, opts)
			if err != nil {
				LogAndFail(res, err, appName)
				return
			}

			for _, entity := range posts {
				post := PostListItem(entity.(editor.Editable), t, status)
				_, err = b.Write(post)
				if err != nil {
					LogAndFail(res, err, appName)
					return
				}
			}
		}

		html += `<ul class="posts row">`

		_, err = b.Write([]byte(`</ul>`))
		if err != nil {
			LogAndFail(res, err, appName)
			return
		}

		statusDisabled := "disabled"
		prevStatus := ""
		nextStatus := ""
		// total may be less than 10 (default count), so reset count to match total
		if total < count {
			count = total
		}
		// nothing previous to current list
		if offset == 0 {
			prevStatus = statusDisabled
		}
		// nothing after current list
		if (offset+1)*count >= total {
			nextStatus = statusDisabled
		}

		// set up pagination values
		urlFmt := req.URL.Path + "?count=%d&offset=%d&&order=%s&status=%s&type=%s"
		prevURL := fmt.Sprintf(urlFmt, count, offset-1, order, status, t)
		nextURL := fmt.Sprintf(urlFmt, count, offset+1, order, status, t)
		start := 1 + count*offset
		end := start + count - 1

		if total < end {
			end = total
		}

		pagination := fmt.Sprintf(`
	<ul class="pagination row">
		<li class="col s2 waves-effect %s"><a href="%s"><i class="material-icons">chevron_left</i></a></li>
		<li class="col s8">%d to %d of %d</li>
		<li class="col s2 waves-effect %s"><a href="%s"><i class="material-icons">chevron_right</i></a></li>
	</ul>
	`, prevStatus, prevURL, start, end, total, nextStatus, nextURL)

		// show indicator that a collection of items will be listed implicitly, but
		// that none are created yet
		if total < 1 {
			pagination = `
		<ul class="pagination row">
			<li class="col s2 waves-effect disabled"><a href="#"><i class="material-icons">chevron_left</i></a></li>
			<li class="col s8">0 to 0 of 0</li>
			<li class="col s2 waves-effect disabled"><a href="#"><i class="material-icons">chevron_right</i></a></li>
		</ul>
		`
		}

		_, err = b.Write([]byte(pagination + `</div></div>`))
		if err != nil {
			LogAndFail(res, err, appName)
			return
		}

		script := `
	<script>
		$(function() {
			var del = $('.quick-delete-post.__ponzu span');
			del.on('click', function(e) {
				if (confirm("[Ponzu] Please confirm:\n\nAre you sure you want to delete this post?\nThis cannot be undone.")) {
					$(e.target).parent().submit();
				}
			});
		});

		// disable link from being clicked if parent is 'disabled'
		$(function() {
			$('ul.pagination li.disabled a').on('click', function(e) {
				e.preventDefault();
			});
		});
	</script>
	`

		btn := `<div class="col s3">
		<a href="/edit?type=` + t + `" class="btn new-post waves-effect waves-light">
			New ` + t + `
		</a>`

		if _, ok := pt.(interfaces.CSVFormattable); ok {
			btn += `<br/>
				<a href="/contents/export?type=` + t + `&format=csv" class="green darken-4 btn export-post waves-effect waves-light">
					<i class="material-icons left">system_update_alt</i>
					CSV
				</a>`
		}

		html += b.String() + script + btn + `</div></div>`

		adminView, err := views.Admin(html, appName)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(adminView)
	}
}
