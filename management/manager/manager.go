package manager

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/bosssauce/ponzu/management/editor"
)

const managerHTML = `
<div class="card editor">
    <form method="post" action="/admin/edit" enctype="multipart/form-data">
		<input type="hidden" name="id" value="{{.ID}}"/>
		<input type="hidden" name="type" value="{{.Kind}}"/>
		<input type="hidden" name="timestamp" class="timestamp __ponzu" value="" />
		<input type="hidden" name="updated" class="updated __ponzu" value="" />
		{{ .Editor }}
	</form>
	<script>
		$(function() {
			// remove all bad chars from all inputs in the form, except file fields
			$('form input:not([type=file]), form textarea').on('blur', function(e) {
				var val = e.target.value;
				e.target.value = replaceBadChars(val);
			});

			var setDefaultTimeAndDate = function($pt, $pd, $ts, $up, unix) {
				var time = getPartialTime(unix),
					date = getPartialDate(unix);

				$pt.val(time);
				$pd.val(date);
				$ts.val(unix);
				$up.val(unix);
			}

			// set time time and date inputs using the hidden timestamp input.
			// if it is empty, set it to now and use that value for time and date
			var publish_time = $('input.__ponzu.time'),
				publish_date = $('input.__ponzu.date'),
				timestamp = $('input.__ponzu.timestamp'),
				updated = $('input.__ponzu.updated'),
				time;

			if (timestamp.val() !== "") {
				time = timestamp.val();
			} else {
				time = (new Date()).getTime();
			}

			setDefaultTimeAndDate(publish_time, publish_date, timestamp, updated, time);
			
		});
	</script>
</div>
`

type manager struct {
	ID     int
	Kind   string
	Editor template.HTML
}

// Manage ...
func Manage(e editor.Editable, typeName string) ([]byte, error) {
	v, err := e.MarshalEditor()
	if err != nil {
		return nil, fmt.Errorf("Couldn't marshal editor for content %T. %s", e, err.Error())
	}

	m := manager{
		ID:     e.ContentID(),
		Kind:   typeName,
		Editor: template.HTML(v),
	}

	// execute html template into buffer for func return val
	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("manager").Parse(managerHTML))
	tmpl.Execute(buf, m)

	return buf.Bytes(), nil
}
