package editor

import (
	"bytes"
	"fmt"
	"html"
	"reflect"
	"strings"
)

type element struct {
	TagName string
	Attrs   map[string]string
	Name    string
	label   string
	data    string
	viewBuf *bytes.Buffer
}

// Input returns the []byte of an <input> HTML element with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Input(fieldName string, p interface{}, attrs map[string]string) []byte {
	e := newElement("input", attrs["label"], fieldName, p, attrs)

	return domElementSelfClose(e)
}

// Textarea returns the []byte of a <textarea> HTML element with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Textarea(fieldName string, p interface{}, attrs map[string]string) []byte {
	e := newElement("textarea", attrs["label"], fieldName, p, attrs)

	return domElement(e)
}

// Timestamp returns the []byte of an <input> HTML element with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Timestamp(fieldName string, p interface{}, attrs map[string]string) []byte {
	var data string
	val := valueFromStructField(fieldName, p)
	if val == "0" {
		data = ""
	} else {
		data = val
	}

	e := &element{
		TagName: "input",
		Attrs:   attrs,
		Name:    tagNameFromStructField(fieldName, p),
		label:   attrs["label"],
		data:    data,
		viewBuf: &bytes.Buffer{},
	}

	return domElementSelfClose(e)
}

// File returns the []byte of a <input type="file"> HTML element with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func File(fieldName string, p interface{}, attrs map[string]string) []byte {
	name := tagNameFromStructField(fieldName, p)
	tmpl :=
		`<div class="file-input ` + name + ` input-field col s12">
			<label class="active">` + attrs["label"] + `</label>
			<div class="file-field input-field">
				<div class="btn">
					<span>Upload</span>
					<input class="upload" type="file">
				</div>
				<div class="file-path-wrapper">
					<input class="file-path validate" placeholder="` + attrs["label"] + `" type="text">
				</div>
			</div>
			<div class="preview"><div class="img-clip"></div></div>			
			<input class="store ` + name + `" type="hidden" name="` + name + `" value="` + valueFromStructField(fieldName, p) + `" />
		</div>`

	script :=
		`<script>
			$(function() {
				var $file = $('.file-input.` + name + `'),
					upload = $file.find('input.upload'),
					store = $file.find('input.store'),
					preview = $file.find('.preview'),
					clip = preview.find('.img-clip'),
					reset = document.createElement('div'),
					img = document.createElement('img'),
					uploadSrc = store.val();
					preview.hide();
				
				// when ` + name + ` input changes (file is selected), remove
				// the 'name' and 'value' attrs from the hidden store input.
				// add the 'name' attr to ` + name + ` input
				upload.on('change', function(e) {
					resetImage();
				});

				if (uploadSrc.length > 0) {
					$(img).attr('src', store.val());
					clip.append(img);
					preview.show();

					$(reset).addClass('reset ` + name + ` btn waves-effect waves-light grey');
					$(reset).html('<i class="material-icons tiny">clear<i>');
					$(reset).on('click', function(e) {
						e.preventDefault();
						preview.animate({"opacity": 0.1}, 200, function() {
							preview.slideUp(250, function() {
								resetImage();
							});
						})
						
					});
					clip.append(reset);
				}

				function resetImage() {
					store.val('');
					store.attr('name', '');
					upload.attr('name', '` + name + `');
					clip.empty();
				}
			});	
		</script>`

	return []byte(tmpl + script)
}

// Richtext returns the []byte of a rich text editor (provided by http://summernote.org/) with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Richtext(fieldName string, p interface{}, attrs map[string]string) []byte {
	// create wrapper for richtext editor, which isolates the editor's css
	iso := []byte(`<div class="iso-texteditor input-field col s12"><label>` + attrs["label"] + `</label>`)
	isoClose := []byte(`</div>`)

	// create the target element for the editor to attach itself
	attrs["class"] = "richtext " + fieldName
	attrs["id"] = "richtext-" + fieldName
	div := &element{
		TagName: "div",
		Attrs:   attrs,
		Name:    "",
		label:   "",
		data:    "",
		viewBuf: &bytes.Buffer{},
	}

	// create a hidden input to store the value from the struct
	val := valueFromStructField(fieldName, p)
	name := tagNameFromStructField(fieldName, p)
	input := `<input type="hidden" name="` + name + `" class="richtext-value ` + fieldName + `" value="` + html.EscapeString(val) + `"/>`

	// build the dom tree for the entire richtext component
	iso = append(iso, domElement(div)...)
	iso = append(iso, []byte(input)...)
	iso = append(iso, isoClose...)

	script := `
	<script>
		$(function() { 
			var _editor = $('.richtext.` + fieldName + `');
			var hidden = $('.richtext-value.` + fieldName + `');

			_editor.materialnote({
				height: 250,
				placeholder: '` + attrs["placeholder"] + `',
				toolbar: [
					['style', ['bold', 'italic', 'underline', 'clear']],
					['font', ['strikethrough', 'superscript', 'subscript']],
					['fontsize', ['fontsize']],
					['color', ['color']],
					['insert', ['link', 'picture', 'video', 'hr']],					
					['para', ['ul', 'ol', 'paragraph']],
					['height', ['height']],
					['misc', ['codeview']]
				],
				// intercept file insertion, upload and insert img with new src
				onImageUpload: function(files) {
					var data = new FormData();
					data.append("file", files[0]);
					$.ajax({
						data: data,
						type: 'POST',
						url: '/admin/edit/upload',
						cache: false,
						contentType: false,
						processData: false,
						success: function(resp) {
							var img = document.createElement('img');
							img.setAttribute('src', resp.data[0].url);
							_editor.materialnote('insertNode', img);
						},
						error: function(xhr, status, err) {
							console.log(status, err);
						}
					})

				}
			});

			// inject content into editor
			if (hidden.val() !== "") {
				_editor.code(hidden.val());
			}

			// update hidden input with encoded value on different events
			_editor.on('materialnote.change', function(e, content, $editable) {
				hidden.val(replaceBadChars(content));			
			});

			_editor.on('materialnote.paste', function(e) {
				hidden.val(replaceBadChars(_editor.code()));			
			});

			// bit of a hack to stop the editor buttons from causing a refresh when clicked 
			$('.note-toolbar').find('button, i, a').on('click', function(e) { e.preventDefault(); });
		});
	</script>`

	return append(iso, []byte(script)...)
}

// Select returns the []byte of a <select> HTML element plus internal <options> with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Select(fieldName string, p interface{}, attrs, options map[string]string) []byte {
	// options are the value attr and the display value, i.e.
	// <option value="{map key}">{map value}</option>

	// find the field value in p to determine if an option is pre-selected
	fieldVal := valueFromStructField(fieldName, p)

	// may need to alloc a buffer, as we will probably loop through options
	// and append the []byte from domElement() called for each option
	attrs["class"] = "browser-default"
	sel := newElement("select", attrs["label"], fieldName, p, attrs)
	var opts []*element

	// provide a call to action for the select element
	cta := &element{
		TagName: "option",
		Attrs:   map[string]string{"disabled": "true", "selected": "true"},
		data:    "Select an option...",
		viewBuf: &bytes.Buffer{},
	}

	// provide a selection reset (will store empty string in db)
	reset := &element{
		TagName: "option",
		Attrs:   map[string]string{"value": ""},
		data:    "None",
		viewBuf: &bytes.Buffer{},
	}

	opts = append(opts, cta, reset)

	for k, v := range options {
		optAttrs := map[string]string{"value": k}
		if k == fieldVal {
			optAttrs["selected"] = "true"
		}
		opt := &element{
			TagName: "option",
			Attrs:   optAttrs,
			data:    v,
			viewBuf: &bytes.Buffer{},
		}

		opts = append(opts, opt)
	}

	return domElementWithChildrenSelect(sel, opts)
}

// Checkbox returns the []byte of a set of <input type="checkbox"> HTML elements
// wrapped in a <div> with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Checkbox(fieldName string, p interface{}, attrs, options map[string]string) []byte {
	attrs["class"] = "input-field col s12"
	div := newElement("div", attrs["label"], fieldName, p, attrs)

	var opts []*element

	// get the pre-checked options if this is already an existing post
	checkedVals := valueFromStructField(fieldName, p)
	checked := strings.Split(checkedVals, "__ponzu")

	i := 0
	for k, v := range options {
		inputAttrs := map[string]string{
			"type":  "checkbox",
			"value": k,
			"id":    strings.Join(strings.Split(v, " "), "-"),
		}

		// check if k is in the pre-checked values and set to checked
		for _, x := range checked {
			if k == x {
				inputAttrs["checked"] = "checked"
			}
		}

		// create a *element manually using the maodified tagNameFromStructFieldMulti
		// func since this is for a multi-value name
		input := &element{
			TagName: "input",
			Attrs:   inputAttrs,
			Name:    tagNameFromStructFieldMulti(fieldName, i, p),
			label:   v,
			data:    "",
			viewBuf: &bytes.Buffer{},
		}

		opts = append(opts, input)
		i++
	}

	return domElementWithChildrenCheckbox(div, opts)
}

// Tags returns the []byte of a tag input (in the style of Materialze 'Chips') with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Tags(fieldName string, p interface{}, attrs map[string]string) []byte {
	name := tagNameFromStructField(fieldName, p)

	// get the saved tags if this is already an existing post
	values := valueFromStructField(fieldName, p)
	tags := strings.Split(values, "__ponzu")

	html := `
	<div class="col s12 __ponzu-tags ` + name + `">
		<label class="active">` + attrs["label"] + ` (Type and press "Enter")</label>
		<div class="chips ` + name + `"></div>
	`

	var initial []string
	i := 0
	for _, tag := range tags {
		tagName := tagNameFromStructFieldMulti(fieldName, i, p)
		html += `<input type="hidden" class="__ponzu-tag ` + tag + `" name=` + tagName + ` value="` + tag + `"/>`
		initial = append(initial, `{tag: '`+tag+`'}`)
		i++
	}

	script := `
	<script>
		$(function() {
			var tags = $('.__ponzu-tags.` + name + `');
			$('.chips.` + name + `').material_chip({
				data: [` + strings.Join(initial, ",") + `],
				secondaryPlaceholder: '+` + name + `'
			});		

			// handle events specific to tags
			var chips = tags.find('.chips');
			
			chips.on('chip.add', function(e, chip) {
				chips.parent().find('.empty-tag').remove();
				
				var input = $('<input>');
				input.attr({
					class: 'tag '+chip.tag,
					name: '` + name + `.'+String(tags.find('input[type=hidden]').length),
					value: chip.tag,
					type: 'hidden'
				});
				
				tags.append(input);
			});

			chips.on('chip.delete', function(e, chip) {
				// convert tag string to class-like selector "some tag" -> ".some.tag"
				var sel = '.__ponzu-tag';
				if (chip.tag.length > 0) {			
					sel += '.';
					sel += chip.tag.split(' ').join('.');
				}
				chips.parent().find(sel).remove();

				// iterate through all hidden tag inputs to re-name them with the correct ` + name + `.index
				var hidden = chips.parent().find('input[type=hidden]');
				
				// if there are no tags, set a blank
				if (hidden.length === 0) {
					var input = $('<input>');
					input.attr({
						class: 'empty-tag',
						name: '` + name + `',
						type: 'hidden'
					});
					
					tags.append(input);
					return;
				}
				
				for (var i = 0; i < hidden.length; i++) {
					$(hidden[i]).attr('name', '` + name + `.'+String(i));
				}
			});
		});
	</script>
	`

	html += `</div>`

	return []byte(html + script)
}

// domElementSelfClose is a special DOM element which is parsed as a
// self-closing tag and thus needs to be created differently
func domElementSelfClose(e *element) []byte {
	e.viewBuf.Write([]byte(`<div class="input-field col s12">`))
	if e.label != "" {
		e.viewBuf.Write([]byte(`<label class="active" for="` + strings.Join(strings.Split(e.label, " "), "-") + `">` + e.label + `</label>`))
	}
	e.viewBuf.Write([]byte(`<` + e.TagName + ` value="`))
	e.viewBuf.Write([]byte(html.EscapeString(e.data) + `" `))

	for attr, value := range e.Attrs {
		e.viewBuf.Write([]byte(attr + `="` + value + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.Name + `"`))
	e.viewBuf.Write([]byte(` />`))

	e.viewBuf.Write([]byte(`</div>`))
	return e.viewBuf.Bytes()
}

// domElementCheckbox is a special DOM element which is parsed as a
// checkbox input tag and thus needs to be created differently
func domElementCheckbox(e *element) []byte {
	e.viewBuf.Write([]byte(`<p class="col s6">`))
	e.viewBuf.Write([]byte(`<` + e.TagName + ` `))

	for attr, value := range e.Attrs {
		e.viewBuf.Write([]byte(attr + `="` + value + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.Name + `"`))
	e.viewBuf.Write([]byte(` /> `))
	if e.label != "" {
		e.viewBuf.Write([]byte(`<label for="` + strings.Join(strings.Split(e.label, " "), "-") + `">` + e.label + `</label>`))
	}
	e.viewBuf.Write([]byte(`</p>`))
	return e.viewBuf.Bytes()
}

// domElement creates a DOM element
func domElement(e *element) []byte {
	e.viewBuf.Write([]byte(`<div class="input-field col s12">`))

	if e.label != "" {
		e.viewBuf.Write([]byte(`<label class="active" for="` + strings.Join(strings.Split(e.label, " "), "-") + `">` + e.label + `</label>`))
	}
	e.viewBuf.Write([]byte(`<` + e.TagName + ` `))

	for attr, value := range e.Attrs {
		e.viewBuf.Write([]byte(attr + `="` + string(value) + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.Name + `"`))
	e.viewBuf.Write([]byte(` >`))

	e.viewBuf.Write([]byte(html.EscapeString(e.data)))
	e.viewBuf.Write([]byte(`</` + e.TagName + `>`))

	e.viewBuf.Write([]byte(`</div>`))
	return e.viewBuf.Bytes()
}

func domElementWithChildrenSelect(e *element, children []*element) []byte {
	e.viewBuf.Write([]byte(`<div class="input-field col s6">`))

	e.viewBuf.Write([]byte(`<` + e.TagName + ` `))

	for attr, value := range e.Attrs {
		e.viewBuf.Write([]byte(attr + `="` + string(value) + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.Name + `"`))
	e.viewBuf.Write([]byte(` >`))

	// loop over children and create domElement for each child
	for _, child := range children {
		e.viewBuf.Write(domElement(child))
	}

	e.viewBuf.Write([]byte(`</` + e.TagName + `>`))

	if e.label != "" {
		e.viewBuf.Write([]byte(`<label class="active">` + e.label + `</label>`))
	}

	e.viewBuf.Write([]byte(`</div>`))
	return e.viewBuf.Bytes()
}

func domElementWithChildrenCheckbox(e *element, children []*element) []byte {
	e.viewBuf.Write([]byte(`<` + e.TagName + ` `))

	for attr, value := range e.Attrs {
		e.viewBuf.Write([]byte(attr + `="` + value + `" `))
	}

	e.viewBuf.Write([]byte(` >`))

	if e.label != "" {
		e.viewBuf.Write([]byte(`<label class="active">` + e.label + `</label>`))
	}

	// loop over children and create domElement for each child
	for _, child := range children {
		e.viewBuf.Write(domElementCheckbox(child))
	}

	e.viewBuf.Write([]byte(`</` + e.TagName + `><div class="clear padding">&nbsp;</div>`))

	return e.viewBuf.Bytes()
}

func tagNameFromStructField(name string, post interface{}) string {
	// sometimes elements in these environments will not have a name,
	// and thus no tag name in the struct which correlates to it.
	if name == "" {
		return name
	}

	field, ok := reflect.TypeOf(post).Elem().FieldByName(name)
	if !ok {
		panic("Couldn't get struct field for: " + name + ". Make sure you pass the right field name to editor field elements.")
	}

	tag, ok := field.Tag.Lookup("json")
	if !ok {
		panic("Couldn't get json struct tag for: " + name + ". Struct fields for content types must have 'json' tags.")
	}

	return tag
}

// due to the format in which gorilla/schema expects form names to be when
// one is associated with multiple values, we need to output the name as such.
// Ex. 'category.0', 'category.1', 'category.2' and so on.
func tagNameFromStructFieldMulti(name string, i int, post interface{}) string {
	tag := tagNameFromStructField(name, post)

	return fmt.Sprintf("%s.%d", tag, i)
}

func valueFromStructField(name string, post interface{}) string {
	field := reflect.Indirect(reflect.ValueOf(post)).FieldByName(name)

	switch field.Kind() {
	case reflect.String:
		return field.String()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%v", field.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%v", field.Uint())

	case reflect.Bool:
		return fmt.Sprintf("%t", field.Bool())

	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%v", field.Complex())

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", field.Float())

	case reflect.Slice:
		s := []string{}

		for i := 0; i < field.Len(); i++ {
			pos := field.Index(i)
			s = append(s, fmt.Sprintf("%v", pos))
		}

		return strings.Join(s, "__ponzu")

	default:
		panic(fmt.Sprintf("Ponzu: Type '%s' for field '%s' not supported.", field.Type(), name))
	}
}

func newElement(tagName, label, fieldName string, p interface{}, attrs map[string]string) *element {
	return &element{
		TagName: tagName,
		Attrs:   attrs,
		Name:    tagNameFromStructField(fieldName, p),
		label:   label,
		data:    valueFromStructField(fieldName, p),
		viewBuf: &bytes.Buffer{},
	}
}
