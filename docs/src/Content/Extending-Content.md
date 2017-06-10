title: Extending Content through built-in Interfaces and optional Addons

Extending your Content types with more features and functionality within the system
is done by implementing the various built-in interfaces provided by Ponzu. To learn 
more about interfaces, see [A Tour of Go - Interfaces](https://tour.golang.org/methods/10).

It is also common to add more advanced functionality to Content types using Addons. Refer to the [Addon documentation](/Ponzu-Addons) for more information about how to use and create Ponzu Addons.

## [Item Interfaces](/Interfaces/Item)

All Content types which embed an `item.Item` will implicitly [implement](#) its many
interfaces. In Ponzu, the following interfaces are exported from the `system/item`
package and have a default implementation which can be overridden to change your
content types' functionality within the system.

- [`item.Pushable`](/Interfaces/Item#itempushable)
- [`item.Hideable`](/Interfaces/Item#itemhideable)
- [`item.Omittable`](/Interfaces/Item#itemomittable)
- [`item.Hookable`](/Interfaces/Item#itemhookable)
- [`item.Identifiable`](/Interfaces/Item#itemidentifiable)
- [`item.Sortable`](/Interfaces/Item#itemsortable)
- [`item.Sluggable`](/Interfaces/Item#itemsluggable)

## [API Interfaces](/Interfaces/API)

To enable 3rd-party clients to interact with your Content types, you can extend your types with the API interfaces:

- [`api.Createable`](/Interfaces/API/#apicreateable)
- [`api.Updateable`](/Interfaces/API/#apiupdateable)
- [`api.Deleteable`](/Interfaces/API/#apideleteable)
- [`api.Trustable`](/Interfaces/API/#apitrustable)

## [Editor Interfaces](/Interfaces/Editor)

To manage how content is edited and handled in the CMS, use the following Editor interfaces:

- [`editor.Editable`](/Interfaces/Editor/#editoreditable)
- [`editor.Mergeable`](/Interfaces/Editor/#editormergeable)

## [Search Interfaces](/Interfaces/Search)

To enable and customize full-text search on your content types, use the following interfaces:

- [`search.Searchable`](/Interfaces/Search/#searchsearchable)