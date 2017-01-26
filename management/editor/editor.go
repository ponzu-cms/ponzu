// Package editor enables users to create edit views from their content
// structs so that admins can manage content
package editor

import (
	"bytes"
	"log"
	"net/http"
)

// Editable ensures data is editable
type Editable interface {
	MarshalEditor() ([]byte, error)
}

// Mergeable allows external post content to be approved and published through
// the public-facing API
type Mergeable interface {
	// Approve copies an external post to the internal collection and triggers
	// a re-sort of its content type posts
	Approve(http.ResponseWriter, *http.Request) error
}

// Editor is a view containing fields to manage content
type Editor struct {
	ViewBuf *bytes.Buffer
}

// Field is used to create the editable view for a field
// within a particular content struct
type Field struct {
	View []byte
}

// Form takes editable content and any number of Field funcs to describe the edit
// page for any content struct added by a user
func Form(post Editable, fields ...Field) ([]byte, error) {
	editor := &Editor{}

	editor.ViewBuf = &bytes.Buffer{}
	_, err := editor.ViewBuf.WriteString(`<table><tbody class="row"><tr class="col s8 editor-fields"><td class="col s12">`)
	if err != nil {
		log.Println("Error writing HTML string to editor Form buffer")
		return nil, err
	}

	for _, f := range fields {
		addFieldToEditorView(editor, f)
	}

	_, err = editor.ViewBuf.WriteString(`</td></tr>`)
	if err != nil {
		log.Println("Error writing HTML string to editor Form buffer")
		return nil, err
	}

	// content items with Item embedded have some default fields we need to render
	_, err = editor.ViewBuf.WriteString(`<tr class="col s4 default-fields"><td class="col s12">`)
	if err != nil {
		log.Println("Error writing HTML string to editor Form buffer")
		return nil, err
	}

	publishTime := `
<div class="row content-only __ponzu">
	<div class="input-field col s6">
		<label class="active">MM</label>
		<select class="month __ponzu browser-default">
			<option value="1">Jan - 01</option>
			<option value="2">Feb - 02</option>
			<option value="3">Mar - 03</option>
			<option value="4">Apr - 04</option>
			<option value="5">May - 05</option>
			<option value="6">Jun - 06</option>
			<option value="7">Jul - 07</option>
			<option value="8">Aug - 08</option>
			<option value="9">Sep - 09</option>
			<option value="10">Oct - 10</option>
			<option value="11">Nov - 11</option>
			<option value="12">Dec - 12</option>
		</select>
	</div>
	<div class="input-field col s2">
		<label class="active">DD</label>
		<input value="" class="day __ponzu" maxlength="2" type="text" placeholder="DD" />
	</div>
	<div class="input-field col s4">
		<label class="active">YYYY</label>
		<input value="" class="year __ponzu" maxlength="4" type="text" placeholder="YYYY" />
	</div>
</div>

<div class="row content-only __ponzu">
	<div class="input-field col s3">
		<label class="active">HH</label>
		<input value="" class="hour __ponzu" maxlength="2" type="text" placeholder="HH" />
	</div>
	<div class="col s1">:</div>
	<div class="input-field col s3">
		<label class="active">MM</label>
		<input value="" class="minute __ponzu" maxlength="2" type="text" placeholder="MM" />
	</div>
	<div class="input-field col s4">
		<label class="active">Period</label>
		<select class="period __ponzu browser-default">
			<option value="AM">AM</option>
			<option value="PM">PM</option>
		</select>
	</div>
</div>
	`

	_, err = editor.ViewBuf.WriteString(publishTime)
	if err != nil {
		log.Println("Error writing HTML string to editor Form buffer")
		return nil, err
	}

	err = addPostDefaultFieldsToEditorView(post, editor)
	if err != nil {
		return nil, err
	}

	submit := `
<div class="input-field post-controls">
	<button class="right waves-effect waves-light btn green save-post" type="submit">Save</button>
	<button class="right waves-effect waves-light btn red delete-post" type="submit">Delete</button>
</div>
`
	_, ok := post.(Mergeable)
	if ok {
		submit +=
			`
<div class="row external post-controls">
	<div class="col s12 input-field">
		<button class="right waves-effect waves-light btn blue approve-post" type="submit">Approve</button>
		<button class="right waves-effect waves-light btn grey darken-2 reject-post" type="submit">Reject</button>
	</div>	
	<label class="approve-details right-align col s12">This content is pending approval. By clicking 'Approve', it will be immediately published. By clicking 'Reject', it will be deleted.</label> 
</div>
`
	}

	script := `
<script>
	$(function() {
		var form = $('form'),
			save = form.find('button.save-post'),
			del = form.find('button.delete-post'),
			external = form.find('.post-controls.external'),
			id = form.find('input[name=id]'),
			timestamp = $('.__ponzu.content-only'),
			slug = $('input[name=slug]');
		
		// hide if this is a new post, or a non-post editor page
		if (id.val() === '-1' || form.attr('action') !== '/admin/edit') {
			del.hide();
			external.hide();
		}

		// hide approval if not on a pending content item
		if (getParam('status') !== 'pending') {
			external.hide();
		} 

		// no timestamp, slug visible on addons
		if (form.attr('action') === '/admin/addon') {
			timestamp.hide();
			slug.parent().hide();
		}

		save.on('click', function(e) {
			e.preventDefault();

			if (getParam('status') === 'pending') {
				var action = form.attr('action');
				form.attr('action', action + '?status=pending')
			}

			form.submit();
		});

		del.on('click', function(e) {
			e.preventDefault();
			var action = form.attr('action');
			action = action + '/delete';
			form.attr('action', action);
			
			if (confirm("[Ponzu] Please confirm:\n\nAre you sure you want to delete this post?\nThis cannot be undone.")) {
				form.submit();
			}
		});

		external.find('button.approve-post').on('click', function(e) {
			e.preventDefault();
			var action = form.attr('action');
			action = action + '/approve';
			form.attr('action', action);

			form.submit();
		});

		external.find('button.reject-post').on('click', function(e) {
			e.preventDefault();
			var action = form.attr('action');
			action = action + '/delete?reject=true';
			form.attr('action', action);

			if (confirm("[Ponzu] Please confirm:\n\nAre you sure you want to reject this post?\nDoing so will delete it, and cannot be undone.")) {
				form.submit();
			}
		});
	});
</script>
`
	_, err = editor.ViewBuf.WriteString(submit + script + `</td></tr></tbody></table>`)
	if err != nil {
		log.Println("Error writing HTML string to editor Form buffer")
		return nil, err
	}

	return editor.ViewBuf.Bytes(), nil
}

func addFieldToEditorView(e *Editor, f Field) error {
	_, err := e.ViewBuf.Write(f.View)
	if err != nil {
		log.Println("Error writing field view to editor view buffer")
		return err
	}

	return nil
}

func addPostDefaultFieldsToEditorView(p Editable, e *Editor) error {
	defaults := []Field{
		{
			View: Input("Slug", p, map[string]string{
				"label":       "URL Slug",
				"type":        "text",
				"disabled":    "true",
				"placeholder": "Will be set automatically",
			}),
		},
		{
			View: Timestamp("Timestamp", p, map[string]string{
				"type":  "hidden",
				"class": "timestamp __ponzu",
			}),
		},
		{
			View: Timestamp("Updated", p, map[string]string{
				"type":  "hidden",
				"class": "updated __ponzu",
			}),
		},
	}

	for _, f := range defaults {
		err := addFieldToEditorView(e, f)
		if err != nil {
			return err
		}
	}

	return nil
}
