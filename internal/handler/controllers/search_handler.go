package controllers

import (
	"bytes"
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/application/search"
	"github.com/fanky5g/ponzu/internal/domain/services/management/editor"
	"github.com/fanky5g/ponzu/internal/handler/controllers/mappers"
	"github.com/fanky5g/ponzu/internal/handler/controllers/views"
	"github.com/fanky5g/ponzu/internal/util"
	"log"
	"net/http"
)

func NewSearchHandler(configService config.Service, searchService search.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		q := req.URL.Query()
		t := q.Get("type")
		status := q.Get("status")

		appName, err := configService.GetAppName()
		if err != nil {
			log.Printf("Failed to get app name: %v\n", appName)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		searchRequest, err := mappers.GetSearchRequest(req.URL.Query())
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusBadRequest)
			errView, err := views.Admin(util.Html("error_400"), appName)
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		if t == "" {
			http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
			return
		}

		matches, err := searchService.Search(t, searchRequest.Query, searchRequest.Count, searchRequest.Offset)
		if err != nil {
			LogAndFail(res, err, appName)
			return
		}

		b := &bytes.Buffer{}
		html := `<div class="col s9 card">
					<div class="card-content">
					<div class="row">
					<div class="card-title col s7">` + t + ` Results</div>
					<form class="col s4" action="/contents/search" method="get">
						<div class="input-field post-search inline">
							<label class="active">Search:</label>
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Within all ` + t + ` fields" class="search"/>
							<input type="hidden" name="type" value="` + t + `" />
							<input type="hidden" name="status" value="` + status + `" />
						</div>
                   </form>
					</div>
					<ul class="posts row">`

		for i := range matches {
			post := PostListItem(matches[i].(editor.Editable), t, status)
			_, err = b.Write(post)
			if err != nil {
				log.Println(err)

				res.WriteHeader(http.StatusInternalServerError)
				errView, err := views.Admin(util.Html("error_500"), appName)
				if err != nil {
					log.Println(err)
				}

				res.Write(errView)
				return
			}
		}

		_, err = b.WriteString(`</ul></div></div>`)
		if err != nil {
			log.Println(err)

			res.WriteHeader(http.StatusInternalServerError)
			errView, err := views.Admin(util.Html("error_500"), appName)
			if err != nil {
				log.Println(err)
			}

			res.Write(errView)
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
