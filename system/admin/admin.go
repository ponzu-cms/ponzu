// Package admin desrcibes the admin view containing references to
// various managers and editors
package admin

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/bosssauce/ponzu/content"
	"github.com/bosssauce/ponzu/system/admin/user"
	"github.com/bosssauce/ponzu/system/db"
)

var startAdminHTML = `<!doctype html>
<html lang="en">
    <head>
        <title>{{ .Logo }}</title>
        <script type="text/javascript" src="/admin/static/common/js/jquery-2.1.4.min.js"></script>
        <script type="text/javascript" src="/admin/static/common/js/util.js"></script>
        <script type="text/javascript" src="/admin/static/dashboard/js/materialize.min.js"></script>
        <script type="text/javascript" src="/admin/static/dashboard/js/chart.bundle.min.js"></script>
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
                <a class="brand-logo" href="/admin">{{ .Logo }}</a>

                <ul class="right">
                    <li><a href="/admin/logout">Logout</a></li>
                </ul>
            </div>
            </nav>
        </div>

        <div class="admin-ui row">`

var mainAdminHTML = `
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
                        <li><a class="col s12" href="/admin/configure/users"><i class="tiny left material-icons">supervisor_account</i>Users</a></li>
                    </div>
                </ul>
                </div>
            </div>
            {{ if .Subview}}
            <div class="subview col s9">
                {{ .Subview }}
            </div>
            {{ end }}`

var endAdminHTML = `
        </div>
        <footer class="row">
            <div class="col s12">
                <p class="center-align">Powered by &copy; <a target="_blank" href="https://ponzu-cms.org">Ponzu</a> &nbsp;&vert;&nbsp; open-sourced by <a target="_blank" href="https://www.bosssauce.it">Boss Sauce Creative</a></p>
            </div>     
        </footer>
    </body>
</html>`

type admin struct {
	Logo    string
	Types   map[string]func() interface{}
	Subview template.HTML
}

// Admin ...
func Admin(view []byte) ([]byte, error) {
	cfg, err := db.Config("name")
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		cfg = []byte("")
	}

	a := admin{
		Logo:    string(cfg),
		Types:   content.Types,
		Subview: template.HTML(view),
	}

	buf := &bytes.Buffer{}
	html := startAdminHTML + mainAdminHTML + endAdminHTML
	tmpl := template.Must(template.New("admin").Parse(html))
	err = tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

var initAdminHTML = `
<div class="init col s5">
<div class="card">
<div class="card-content">
    <div class="card-title">Welcome!</div>
    <blockquote>You need to initialize your system by filling out the form below. All of 
    this information can be updated later on, but you will not be able to start 
    without first completing this step.</blockquote>
    <form method="post" action="/admin/init" class="row">
        <div>Configuration</div>
        <div class="input-field col s12">        
            <input placeholder="Enter the name of your site (interal use only)" class="validate required" type="text" id="name" name="name"/>
            <label for="name" class="active">Site Name</label>
        </div>
        <div class="input-field col s12">        
            <input placeholder="Used for acquiring SSL certificate (e.g. www.example.com or  example.com)" class="validate" type="text" id="domain" name="domain"/>
            <label for="domain" class="active">Domain</label>
        </div>
        <div>Admin Details</div>
        <div class="input-field col s12">
            <input placeholder="Your email address e.g. you@example.com" class="validate required" type="email" id="email" name="email"/>
            <label for="email" class="active">Email</label>
        </div>
        <div class="input-field col s12">
            <input placeholder="Enter a strong password" class="validate required" type="password" id="password" name="password"/>
            <label for="password" class="active">Password</label>        
        </div>
        <button class="btn waves-effect waves-light right">Start</button>
    </form>
</div>
</div>
</div>
<script>
    $(function() {
        $('.nav-wrapper ul.right').hide();
        
        var logo = $('a.brand-logo');
        var name = $('input#name');

        name.on('change', function(e) {
            logo.text(e.target.value);
        });
    });
</script>
`

// Init ...
func Init() ([]byte, error) {
	html := startAdminHTML + initAdminHTML + endAdminHTML

	cfg, err := db.Config("name")
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		cfg = []byte("")
	}

	a := admin{
		Logo: string(cfg),
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("init").Parse(html))
	err = tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

var loginAdminHTML = `
<div class="init col s5">
<div class="card">
<div class="card-content">
    <div class="card-title">Welcome!</div>
    <blockquote>Please log in to the system using your email address and password.</blockquote>
    <form method="post" action="/admin/login" class="row">
        <div class="input-field col s12">
            <input placeholder="Enter your email address e.g. you@example.com" class="validate required" type="email" id="email" name="email"/>
            <label for="email" class="active">Email</label>
        </div>
        <div class="input-field col s12">
            <input placeholder="Enter your password" class="validate required" type="password" id="password" name="password"/>
            <label for="password" class="active">Password</label>        
        </div>
        <button class="btn waves-effect waves-light right">Log in</button>
    </form>
</div>
</div>
</div>
<script>
    $(function() {
        $('.nav-wrapper ul.right').hide();
    });
</script>
`

// Login ...
func Login() ([]byte, error) {
	html := startAdminHTML + loginAdminHTML + endAdminHTML

	cfg, err := db.Config("name")
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		cfg = []byte("")
	}

	a := admin{
		Logo: string(cfg),
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("login").Parse(html))
	err = tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UsersList ...
func UsersList(req *http.Request) ([]byte, error) {
	html := `
    <div class="card user-management">
        <div class="card-title">Edit your account:</div>    
        <form class="row" enctype="multipart/form-data" action="/admin/configure/users/edit" method="post">
            <div class="input-feild col s9">
                <label class="active">Email Address</label>
                <input type="email" name="email" value="{{ .User.Email }}"/>
            </div>

            <div class="input-feild col s9">
                <div>To approve changes, enter your password:</div>
                
                <label class="active">Current Password</label>
                <input type="password" name="password"/>
            </div>

            <div class="input-feild col s9">
                <label class="active">New Password: (leave blank if no password change needed)</label>
                <input name="new_password" type="password"/>
            </div>

            <div class="input-feild col s9">                        
                <button class="btn waves-effect waves-light green right" type="submit">Save</button>
            </div>
        </form>

        <div class="card-title">Add a new user:</div>        
        <form class="row" enctype="multipart/form-data" action="/admin/configure/users" method="post">
            <div class="input-feild col s9">
                <label class="active">Email Address</label>
                <input type="email" name="email" value=""/>
            </div>

            <div class="input-feild col s9">
                <label class="active">Password</label>
                <input type="password" name="password"/>
            </div>

            <div class="input-feild col s9">            
                <button class="btn waves-effect waves-light green right" type="submit">Add User</button>
            </div>   
        </form>        

        <div class="card-title">Remove Admin Users</div>        
        <ul class="users row">
            {{ range .Users }}
            <li class="col s9">
                {{ .Email }}
                <form enctype="multipart/form-data" class="delete-user __ponzu right" action="/admin/configure/users/delete" method="post">
                    <span>Delete</span>
                    <input type="hidden" name="email" value="{{ .Email }}"/>
                    <input type="hidden" name="id" value="{{ .ID }}"/>
                </form>
            </li>
            {{ end }}
        </ul>
    </div>
    `
	script := `
    <script>
        $(function() {
            var del = $('.delete-user.__ponzu span');
            del.on('click', function(e) {
                if (confirm("[Ponzu] Please confirm:\n\nAre you sure you want to delete this user?\nThis cannot be undone.")) {
                    $(e.target).parent().submit();
                }
            });
        });
    </script>
    `
	// get current user out to pass as data to execute template
	j, err := db.CurrentUser(req)
	if err != nil {
		return nil, err
	}

	var usr user.User
	err = json.Unmarshal(j, &usr)
	if err != nil {
		return nil, err
	}

	// get all users to list
	jj, err := db.UserAll()
	if err != nil {
		return nil, err
	}

	var usrs []user.User
	for i := range jj {
		var u user.User
		err = json.Unmarshal(jj[i], &u)
		if err != nil {
			return nil, err
		}
		if u.Email != usr.Email {
			usrs = append(usrs, u)
		}
	}

	// make buffer to execute html into then pass buffer's bytes to Admin
	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("users").Parse(html + script))
	data := map[string]interface{}{
		"User":  usr,
		"Users": usrs,
	}

	err = tmpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}

	view, err := Admin(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return view, nil
}

var analyticsHTML = `
<div class="analytics">
<div class="card">
<div class="card-content">
    <div class="card-title">API Requests</div>
    <canvas id="analytics-chart" width="100%" height="300px"></canvas>
    <script>
    var ctx = document.getElementById("analytics-chart");
    var myChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: ["10/28", "10/29", "10/30", "10/31", "11/1", "11/2", "11/3"],
            datasets: [{
                label: 'Total Requests by Day',
                data: [12332, 19333, 13545, 51776, 22334, 13334, 9089],
                backgroundColor: 'rgba(54, 162, 235, 0.2)',
                borderColor: 'rgba(54, 162, 235, 1)',
                borderWidth: 1
            }]
        },
        options: {
            scales: {
                yAxes: [{
                    ticks: {
                        beginAtZero:true
                    }
                }]
            }
        }
    });
    </script>
</div>
</div>
</div>
`

var err400HTML = `
<div class="error-page e400 col s6">
<div class="card">
<div class="card-content">
    <div class="card-title"><b>400</b> Error: Bad Request</div>
    <blockquote>Sorry, the request was unable to be completed.</blockquote>
</div>
</div>
</div>
`

// Error400 creates a subview for a 400 error page
func Error400() ([]byte, error) {
	view, err := Admin([]byte(err400HTML))
	if err != nil {
		return nil, err
	}

	return view, nil
}

var err404HTML = `
<div class="error-page e404 col s6">
<div class="card">
<div class="card-content">
    <div class="card-title"><b>404</b> Error: Not Found</div>
    <blockquote>Sorry, the page you requested could not be found.</blockquote>
</div>
</div>
</div>
`

// Error404 creates a subview for a 404 error page
func Error404() ([]byte, error) {
	view, err := Admin([]byte(err404HTML))
	if err != nil {
		return nil, err
	}

	return view, nil
}

var err405HTML = `
<div class="error-page e405 col s6">
<div class="card">
<div class="card-content">
    <div class="card-title"><b>405</b> Error: Method Not Allowed</div>
    <blockquote>Sorry, the method of your request is not allowed.</blockquote>
</div>
</div>
</div>
`

// Error405 creates a subview for a 405 error page
func Error405() ([]byte, error) {
	view, err := Admin([]byte(err405HTML))
	if err != nil {
		return nil, err
	}

	return view, nil
}

var err500HTML = `
<div class="error-page e500 col s6">
<div class="card">
<div class="card-content">
    <div class="card-title"><b>500</b> Error: Internal Service Error</div>
    <blockquote>Sorry, something unexpectedly went wrong.</blockquote>
</div>
</div>
</div>
`

// Error500 creates a subview for a 500 error page
func Error500() ([]byte, error) {
	view, err := Admin([]byte(err500HTML))
	if err != nil {
		return nil, err
	}

	return view, nil
}
