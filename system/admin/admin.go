// Package admin desrcibes the admin view containing references to
// various managers and editors
package admin

import (
	"bytes"
	"html/template"

	"github.com/nilslice/cms/content"
)

const adminHTML = `<!doctype html>
<html lang="en">
    <head>
        <title>CMS</title>
        <script type="text/javascript" src="/admin/static/common/js/jquery-2.1.4.min.js"></script>
        <script type="text/javascript" src="/admin/static/common/js/base64.min.js"></script>
        <script type="text/javascript" src="/admin/static/common/js/util.js"></script>
        <script type="text/javascript" src="/admin/static/dashboard/js/materialize.min.js"></script>
        <script type="text/javascript" src="/admin/static/editor/js/materialNote.js"></script> 
        <script type="text/javascript" src="/admin/static/editor/js/ckMaterializeOverrides.js"></script>
                  
        <link rel="stylesheet" href="/admin/static/dashboard/css/material-icons.css" />     
        <link rel="stylesheet" href="/admin/static/dashboard/css/materialize.min.css" />
        <link rel="stylesheet" href="/admin/static/editor/css/materialNote.css" />
        <link rel="stylesheet" href="/admin/static/dashboard/css/admin.css" />    

        <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
        <meta charset="utf-8">
        <meta http-equiv="X-UA-Compatible" content="IE=edge">
    </head>
    <body class="grey lighten-4">
       <div class="navbar-fixed">
            <nav class="grey darken-2">
            <div class="nav-wrapper">
                <a class="brand-logo" href="/admin">CMS</a>

                <ul class="right">
                    <li><a href="/admin/logout">Logout</a></li>
                </ul>
            </div>
            </nav>
        </div>

        <div class="admin-ui row">
            
            <div class="left-nav col s3">
                <div class="card">
                <ul class="card-content collection">
                    <div class="card-title">Content</div>
                                    
                    {{ range $t, $f := .Types }}
                    <div class="row collection-item">
                        <li><a class="col s12" href="/admin/posts?type={{ $t }}"><i class="tiny left material-icons">playlist_add</i>{{ $t }}</a></li>
                    </div>
                    {{ end }}

                    <div class="card-title">System</div>                                
                    <div class="row collection-item">
                        <li><a class="col s12" href="/admin/configure"><i class="tiny left material-icons">settings</i>Configuration</a></li>
                    </div>
                </ul>
                </div>
            </div>
            {{ if .Subview}}
            <div class="subview col s9">
                {{ .Subview }}
            </div>
            {{ end }}
        </div>
    </body>
</html>`

type admin struct {
	Types   map[string]func() interface{}
	Subview template.HTML
}

// Admin ...
func Admin(view []byte) ([]byte, error) {
	a := admin{
		Types:   content.Types,
		Subview: template.HTML(view),
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("admin").Parse(adminHTML))
	err := tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
