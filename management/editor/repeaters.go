package editor

import (
	"bytes"
	"strings"
)

// RepeatSelect returns the []byte of a <select> HTML element plus internal <options> with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func RepeatSelect(fieldName string, p interface{}, attrs, options map[string]string) []byte {
	// options are the value attr and the display value, i.e.
	// <option value="{map key}">{map value}</option>
	scope := tagNameFromStructField(fieldName, p)
	html := &bytes.Buffer{}
	html.WriteString(`<span class="__ponzu-repeat ` + scope + `">`)

	// find the field values in p to determine if an option is pre-selected
	fieldVals := valueFromStructField(fieldName, p)
	vals := strings.Split(fieldVals, "__ponzu")

	attrs["class"] = "browser-default"

	// loop through vals and create selects and options for each, adding to html
	if len(vals) > 0 {
		for i, val := range vals {
			sel := &element{
				TagName: "select",
				Attrs:   attrs,
				Name:    tagNameFromStructFieldMulti(fieldName, i, p),
				viewBuf: &bytes.Buffer{},
			}

			// only add the label to the first select in repeated list
			if i == 0 {
				sel.label = attrs["label"]
			}

			// create options for select element
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
				if k == val {
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

			html.Write(domElementWithChildrenSelect(sel, opts))
		}
	}

	html.WriteString(`</span>`)
	return append(html.Bytes(), RepeatController(fieldName, p, "select")...)
}

// RepeatController generates the javascript to control any repeatable form
// element in an editor based on its type, field name and HTML tag name
func RepeatController(fieldName string, p interface{}, htmlTagName string) []byte {
	scope := tagNameFromStructField(fieldName, p)
	script := `
    <script>
        $(function() {
            // define the scope of the repeater
            var scope = $('.__ponzu-repeat.` + scope + `');

            var getChildren = function() {
                return scope.find('` + htmlTagName + `')
            }

            var resetFieldNames = function() {
                // loop through children, set its name to the fieldName.i where
                // i is the current index number of children array
                var children = getChildren();
                for (var i = 0; i < children.length; i++) {
                    var el = children[i];
                    $(el).attr('name', '` + scope + `.'+String(i));
                    
                    // reset controllers
                    $(el).find('.controls').remove();
                }

                applyRepeatControllers();
            }

            var addRepeater = function(e) {
                e.preventDefault();
                
                // find and clone the repeatable input-like element
                var controls = $(e.target).parent();
                var clone = controls.parent().clone();

                // if repeat has label, remove it
                clone.find('label').remove();
                
                // remove the pre-filled value from clone
                clone.find('` + htmlTagName + `').val("");

                // remove controls if already present
                clone.find('.controls').remove();

                // add clone to scope and reset field name attributes
                scope.append(clone);

                resetFieldNames();
            }

            var delRepeater = function(e) {
                e.preventDefault();

                // do nothing if the input is the only child
                var children = getChildren();
                if (children.length === 1) {
                    return;
                }

                var del = e.target;
                
                // pass label onto next input-like element if del 0 index
                var wrapper = $(del).parent().parent();
                if (wrapper.find('` + htmlTagName + `').attr('name') === '` + scope + `.0') {
                    wrapper.next().append(wrapper.find('label'))
                }
                
                // find the outermost element which is the input-like element
                // within the scope that contains the repeater control (-) and
                // delete it
                $(del).parent().parent().remove();

                resetFieldNames();
            }

            var createControls = function() {
                // create + / - controls for each input-like child element of scope
                var add = $('<button>+</button>');
                add.addClass('repeater-add');
                add.addClass('btn-flat waves-effect waves-green');

                var del = $('<button>-</button>');
                del.addClass('repeater-del');
                del.addClass('btn-flat waves-effect waves-red');                

                var controls = $('<span></span>');
                controls.addClass('controls');
                controls.addClass('right');

                // bind listeners to child's controls
                add.on('click', addRepeater);
                del.on('click', delRepeater);

                controls.append(add);
                controls.append(del);

                return controls;
            }

            var applyRepeatControllers = function() {
                // add controls to each child
                var children = getChildren()
                for (var i = 0; i < children.length; i++) {
                    var el = children[i];
                    
                    $(el).parent().find('.controls').remove();
                    
                    var controls = createControls();                                        
                    $(el).parent().append(controls);
                }
            }

            applyRepeatControllers();
        });

    </script>
    `

	return []byte(script)
}
