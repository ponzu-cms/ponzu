package config

import (
	"github.com/bosssauce/ponzu/content"
	"github.com/bosssauce/ponzu/management/editor"
)

//Config represents the confirgurable options of the system
type Config struct {
	content.Item
	editor editor.Editor

	Name            string   `json:"name"`
	Domain          string   `json:"domain"`
	ClientSecret    string   `json:"client_secret"`
	Etag            string   `json:"etag"`
	CacheInvalidate []string `json:"-"`
}

// SetContentID partially implements editor.Editable
func (c *Config) SetContentID(id int) { c.ID = id }

// ContentID partially implements editor.Editable
func (c *Config) ContentID() int { return c.ID }

// ContentName partially implements editor.Editable
func (c *Config) ContentName() string { return c.Name }

// SetSlug partially implements editor.Editable
func (c *Config) SetSlug(slug string) { c.Slug = slug }

// Editor partially implements editor.Editable
func (c *Config) Editor() *editor.Editor { return &c.editor }

// MarshalEditor writes a buffer of html to edit a Post and partially implements editor.Editable
func (c *Config) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(c,
		editor.Field{
			View: editor.Input("Name", c, map[string]string{
				"label":       "Site Name",
				"placeholder": "Add a name to this site (internal use only)",
			}),
		},
		editor.Field{
			View: editor.Input("Domain", c, map[string]string{
				"label":       "Domain Name (required for SSL certificate)",
				"placeholder": "e.g. www.example.com or example.com",
			}),
		},
		editor.Field{
			View: editor.Input("ClientSecret", c, map[string]string{
				"label":    "Client Secret (used to validate requests)",
				"disabled": "true",
			}),
		},
		editor.Field{
			View: editor.Input("ClientSecret", c, map[string]string{
				"type": "hidden",
			}),
		},
		editor.Field{
			View: editor.Input("Etag", c, map[string]string{
				"label":    "Etag Header (used for static asset cache)",
				"disabled": "true",
			}),
		},
		editor.Field{
			View: editor.Input("Etag", c, map[string]string{
				"type": "hidden",
			}),
		},
		editor.Field{
			View: editor.Checkbox("CacheInvalidate", c, map[string]string{
				"label": "Invalidate cache on save",
			}, map[string]string{
				"cache": "invalidate",
			}),
		},
	)
	if err != nil {
		return nil, err
	}

	open := []byte(`<div class="card"><form action="/admin/configure" method="post">`)
	close := []byte(`</form></div>`)
	script := []byte(`
	<script>
		$(function() {
			// hide default fields & labels unecessary for the config
			var fields = $('.default-fields');
			fields.css('position', 'relative');
			fields.find('input:not([type=submit])').remove();
			fields.find('label').remove();
			fields.find('button').css({
				position: 'absolute',
				top: '-10px',
				right: '0px'
			});

			// adjust layout of td so save button is in same location as usual
			fields.find('td').css('float', 'right');

			// stop some fixed config settings from being modified
			fields.find('input[name=client_secret]').attr('name', '');
		});
	</script>
	`)

	view = append(open, view...)
	view = append(view, close...)
	view = append(view, script...)

	return view, nil
}
