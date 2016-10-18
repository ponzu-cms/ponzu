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

			var updateTimestamp = function(dt, $ts) {
				var year = dt.year.val(),
					month = dt.month.val()-1,
					day = dt.day.val(),
					hour = dt.hour.val(),
					minutes = dt.minute.val();

					if (dt.period == "PM") {
						hours = hours + 12;
					}

				var date = new Date(year, month, day, hour, minute);
				
				$ts.val(date.getTime());
			}

			var setDefaultTimeAndDate = function(dt, $ts, $up, unix) {
				var time = getPartialTime(unix),
					date = getPartialDate(unix);

				dt.hour.val(time.hh);
				dt.minute.val(time.mm);
				dt.period.val(time.pd);
				dt.year.val(date.yyyy);
				dt.month.val(date.mm);
				dt.day.val(date.dd);
				
				$ts.val(unix);					
				$up.val(unix);
			}

			// set time time and date inputs using the hidden timestamp input.
			// if it is empty, set it to now and use that value for time and date
			var publish_time_hh = $('input.__ponzu.hour'),
				publish_time_mm = $('input.__ponzu.minute'),
				publish_time_pd = $('select.__ponzu.period'),				
				publish_date_yyyy = $('input.__ponzu.year'),
				publish_date_mm = $('select.__ponzu.month'),
				publish_date_dd = $('input.__ponzu.day'),
				timestamp = $('input.__ponzu.timestamp'),
				updated = $('input.__ponzu.updated'),
				getFields = function() {
					return {
						hour: publish_time_hh,
						minute: publish_time_mm,
						period: publish_time_pd,
						year: publish_date_yyyy,
						month: publish_date_mm,
						day: publish_date_dd
					}
				},
				time;

			if (timestamp.val() !== "") {
				time = timestamp.val();
			} else {
				time = (new Date()).getTime();
			}

			setDefaultTimeAndDate(getFields(), timestamp, updated, time);
			
			var timeUpdated = false;
			$('form').on('submit', function(e) {
				if (timeUpdated === true) {
					timeUpdated = false;
					return;
				}

				e.preventDefault();

				updateTimestamp(getFields(), timestamp);

				timeUpdated = true;
				$('form').submit();				
			});
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
