# spreadsheet-cms

a hack to generate static content from a google spreadsheet

## expectations for data

- first row defines fields
- must contain unique id column
- text delimiter is `"`
- field separator `,`
- localize fields, by appending `-<lang>`

[example/data.csv](example/data.csv)

### lists in data

If you want to define a list in your data, the list item delimiter is an new line and use the List function.

## Usage

```bash
spreadsheet-cms -csv https://server.com/data.csv -languages de,en -out path/to/out-dir -asset-dir path/to/assets
```

## Templating

- uses go std lib templates [https://golang.org/pkg/html/template/](https://golang.org/pkg/html/template/)
- the hugo.io docs have a solid intro [https://gohugo.io/templates/introduction/](https://gohugo.io/templates/introduction/)

### Template funcs

#### Check for assets

```html
{{ $asset := print .id ".jpg" }}
{{ if HasAsset $asset }}
    <img src="./path/to/assets/{{$asset}}">
{{ end }}
```

#### Check for empty values

```html
{{ if Empty .name }}
    <h1>No name given</h1>
{{ else }}
    <h1>{{.name}}</h1>
{{ end }}
```

#### Lists

```html
<ul>
    {{ range $listItem := List .features }}
        <li>{{$listItem}}</li>
    {{ end }}
</ul>
```
