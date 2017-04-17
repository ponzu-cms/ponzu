// Package manager contains the admin UI to the CMS which wraps all content editor
// interfaces to manage the create/edit/delete capabilities of Ponzu content.
package manager

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/ponzu-cms/ponzu/management/editor"
	"github.com/ponzu-cms/ponzu/system/item"

	uuid "github.com/satori/go.uuid"
)

const managerHTML = `
<div class="card editor">
    <form method="post" action="/admin/edit" enctype="multipart/form-data">
		<input type="hidden" name="uuid" value="{{.UUID}}"/>
		<input type="hidden" name="id" value="{{.ID}}"/>
		<input type="hidden" name="type" value="{{.Kind}}"/>
		<input type="hidden" name="slug" value="{{.Slug}}"/>
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
				var year = parseInt(dt.year.val()),
					month = parseInt(dt.month.val())-1,
					day = parseInt(dt.day.val()),
					hour = parseInt(dt.hour.val()),
					minute = parseInt(dt.minute.val());

					if (dt.period.val() === "PM") {
						hour = hour + 12;
					}

				// add seconds to Date() to differentiate times more precisely,
				// although not 100% accurately
				var sec = (new Date()).getSeconds();
				var date = new Date(year, month, day, hour, minute, sec);
				
				$ts.val(date.getTime());
			}

			var setDefaultTimeAndDate = function(dt, unix) {
				var time = getPartialTime(unix),
					date = getPartialDate(unix);

				dt.hour.val(time.hh);
				dt.minute.val(time.mm);
				dt.period.val(time.pd);
				dt.year.val(date.yyyy);
				dt.month.val(date.mm);
				dt.day.val(date.dd);				
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
				time = parseInt(timestamp.val());
			} else {
				time = (new Date()).getTime();
			}

			setDefaultTimeAndDate(getFields(), time);
			
			var timeUpdated = false;
			$('form').on('submit', function(e) {
				if (timeUpdated === true) {
					timeUpdated = false;
					return;
				}

				e.preventDefault();

				updateTimestamp(getFields(), timestamp);
				updated.val((new Date()).getTime());

				timeUpdated = true;
				$('form').submit();
			});
		});

	</script>
</div>
`

var managerTmpl = template.Must(template.New("manager").Parse(managerHTML))

type manager struct {
	ID     int
	UUID   uuid.UUID
	Kind   string
	Slug   string
	Editor template.HTML
}

// Manage ...
func Manage(e editor.Editable, typeName string) ([]byte, error) {
	v, err := e.MarshalEditor()
	if err != nil {
		return nil, fmt.Errorf("Couldn't marshal editor for content %s. %s", typeName, err.Error())
	}

	i, ok := e.(item.Identifiable)
	if !ok {
		return nil, fmt.Errorf("Content type %s does not implement item.Identifiable.", typeName)
	}

	s, ok := e.(item.Sluggable)
	if !ok {
		return nil, fmt.Errorf("Content type %s does not implement item.Sluggable.", typeName)
	}

	m := manager{
		ID:     i.ItemID(),
		UUID:   i.UniqueID(),
		Kind:   typeName,
		Slug:   s.ItemSlug(),
		Editor: template.HTML(v),
	}

	// execute html template into buffer for func return val
	buf := &bytes.Buffer{}
	if err := managerTmpl.Execute(buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
