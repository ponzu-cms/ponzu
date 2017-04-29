package item

import (
	"fmt"

	"github.com/ponzu-cms/ponzu/management/editor"
)

// FileUpload represents the file uploaded to the system
type FileUpload struct {
	Item

	Name          string `json:"name"`
	Path          string `json:"path"`
	ContentLength int64  `json:"content_length"`
	ContentType   string `json:"content_type"`
}

// String partially implements item.Identifiable and overrides Item's String()
func (f *FileUpload) String() string { return f.Name }

// MarshalEditor writes a buffer of html to edit a Post and partially implements editor.Editable
func (f *FileUpload) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(f,
		editor.Field{
			View: func() []byte {
				if f.Path == "" {
					return nil
				}

				return []byte(`
            <div class="input-field col s12">
                <label class="active">` + f.Name + `</label>
                <!-- Add your custom editor field view here. -->
				<h4>` + f.Name + `</h4>

				<img class="preview" src="` + f.Path + `"/>
				<p>File information:</p>
				<ul>
					<li>Content-Length: ` + fmt.Sprintf("%s", FmtBytes(float64(f.ContentLength))) + `</li>
					<li>Content-Type: ` + f.ContentType + `</li>
				</ul>
            </div>
            `)
			}(),
		},
		editor.Field{
			View: editor.File("Path", f, map[string]string{
				"label":       "File Upload",
				"placeholder": "Upload the file here",
			}),
		},
	)
	if err != nil {
		return nil, err
	}

	open := []byte(`
	<div class="card">
		<div class="card-content">
			<div class="card-title">File Uploads</div>
		</div>
		<form action="/admin/uploads" method="post">
	`)
	close := []byte(`</form></div>`)
	script := []byte(`
	<script>
		$(function() {
			// hide default fields & labels unnecessary for the config
			var fields = $('.default-fields');
			fields.css('position', 'relative');
			fields.find('input:not([type=submit])').remove();
			fields.find('label').remove();
			fields.find('button').css({
				position: 'absolute',
				top: '-10px',
				right: '0px'
			});

			var contentOnly = $('.content-only.__ponzu');
			contentOnly.hide();
			contentOnly.find('input, textarea, select').attr('name', '');

			// adjust layout of td so save button is in same location as usual
			fields.find('td').css('float', 'right');

			// stop some fixed config settings from being modified
			fields.find('input[name=client_secret]').attr('name', '');

			// hide save, show delete
			fields.find('.save-post').hide();
			fields.find('.delete-post').show();
		});
	</script>
	`)

	view = append(open, view...)
	view = append(view, close...)
	view = append(view, script...)

	return view, nil
}

func (f *FileUpload) Push() []string {
	return []string{
		"path",
	}
}

// FmtBytes converts the numeric byte size value to the appropriate magnitude
// size in KB, MB, GB, TB, PB, or EB.
func FmtBytes(size float64) string {
	unit := float64(1024)
	BYTE := unit
	KBYTE := BYTE * unit
	MBYTE := KBYTE * unit
	GBYTE := MBYTE * unit
	TBYTE := GBYTE * unit
	PBYTE := TBYTE * unit

	switch {
	case size < BYTE:
		return fmt.Sprintf("%0.f B", size)
	case size < KBYTE:
		return fmt.Sprintf("%.1f KB", size/BYTE)
	case size < MBYTE:
		return fmt.Sprintf("%.1f MB", size/KBYTE)
	case size < GBYTE:
		return fmt.Sprintf("%.1f GB", size/MBYTE)
	case size < TBYTE:
		return fmt.Sprintf("%.1f TB", size/GBYTE)
	case size < PBYTE:
		return fmt.Sprintf("%.1f PB", size/TBYTE)
	default:
		return fmt.Sprintf("%0.f B", size)
	}

}
