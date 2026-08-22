package report

const html = `
<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="utf-8">
<title>Отчёт о сборке OpenAPI</title>
<style>
body { 
font-family: sans-serif; 
margin: 2em; 
color: #222; 
}
h1 { 
font-size: 1.4em; 
}
h2 { 
font-size: 1.1em; 
margin-top: 1.5em; 
}
table { 
border-collapse: collapse; 
margin-top: 0.5em; 
}
th, td { 
border: 1px solid #ccc; 
padding: 4px 10px; 
text-align: left; 
}
th { 
background: #f0f0f0; 
}
.empty {
color: #888; 
}
.conflict { 
color: #b00020; 
}
</style>
</head>
<body>
<h1>Отчёт о сборке OpenAPI</h1>

<h2>Сводка</h2>
<table>
<tr><th>Параметр</th><th>Значение</th></tr>
<tr><td>base</td><td>{{.Params.Base}}</td></tr>
<tr><td>source</td><td>{{.Params.Source}}</td></tr>
<tr><td>format</td><td>{{.Params.Format}}</td></tr>
<tr><td>strict</td><td>{{.Params.Strict}}</td></tr>
<tr><td>Go-файлов</td><td>{{.Totals.GoFiles}}</td></tr>
<tr><td>Фрагментов</td><td>{{.Totals.Fragments}}</td></tr>
<tr><td>Операций</td><td>{{.Totals.Operations}}</td></tr>
<tr><td>Компонентов</td><td>{{.Totals.Components}}</td></tr>
<tr><td>Тегов</td><td>{{.Totals.Tags}}</td></tr>
<tr><td>Конфликтов</td><td>{{.Totals.Conflicts}}</td></tr>
<tr><td>Предупреждений</td><td>{{.Totals.Warnings}}</td></tr>
</table>

<h2>Операции</h2>
{{if .Operations}}
<table>
<tr><th>Метод</th><th>Путь</th><th>Файл-владелец</th></tr>
{{range .Operations}}
<tr><td>{{.Method}}</td><td>{{.Path}}</td><td>{{.Origin.File}}:{{.Origin.Line}}</td></tr>
{{end}}
</table>
{{else}}<p class="empty">Нет операций.</p>{{end}}

<h2>Компоненты</h2>
{{if .Components}}
<table>
<tr><th>Раздел</th><th>Имя</th><th>Файл-владелец</th></tr>
{{range .Components}}
<tr><td>{{.Section}}</td><td>{{.Name}}</td><td>{{.Origin.File}}:{{.Origin.Line}}</td></tr>
{{end}}
</table>
{{else}}<p class="empty">Нет компонентов.</p>{{end}}

<h2>Теги</h2>
{{if .Tags}}
<ul>{{range .Tags}}<li>{{.}}</li>{{end}}</ul>
{{else}}<p class="empty">Нет тегов.</p>{{end}}

<h2>Конфликты</h2>
{{if .Conflicts}}
<table>
<tr><th>Вид</th><th>Ключ</th><th>Первый источник</th><th>Второй источник</th></tr>
{{range .Conflicts}}
<tr class="conflict"><td>{{.Kind}}</td><td>{{.Key}}</td>
<td>{{.First.File}}:{{.First.Line}}</td><td>{{.Second.File}}:{{.Second.Line}}</td></tr>
{{end}}
</table>
{{else}}<p class="empty">Конфликтов нет.</p>{{end}}

<h2>Предупреждения</h2>
{{if .Warnings}}
<ul>{{range .Warnings}}<li>{{.Kind}}: {{.Target}}</li>{{end}}</ul>
{{else}}<p class="empty">Предупреждений нет.</p>{{end}}

<h2>Ошибки</h2>
{{if .Errors}}
<ul>{{range .Errors}}<li class="conflict">{{.}}</li>{{end}}</ul>
{{else}}<p class="empty">Ошибок нет.</p>{{end}}

</body>
</html>
`
