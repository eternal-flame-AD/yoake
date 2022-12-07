package ui

var headTagDayjs = []string{
	`<script src="https://cdn.jsdelivr.net/npm/dayjs@1.11.6/dayjs.min.js"
	integrity="sha256-EfJOqCcshFS/2TxhArURu3Wn8b/XDA4fbPWKSwZ+1B8=" crossorigin="anonymous"></script>`,
	`    <script src="https://cdn.jsdelivr.net/npm/dayjs@1.11.6/plugin/relativeTime.js"
	integrity="sha256-muryXOPFkVJcJO1YFmhuKyXYmGDT2TYVxivG0MCgRzg=" crossorigin="anonymous"></script>`,
	`    <script src="https://cdn.jsdelivr.net/npm/dayjs@1.11.6/plugin/localizedFormat.js"
	integrity="sha256-g+gxm1xmRq4IecSRujv2eKyUCo/i1b5kRnWNcSbYEO0=" crossorigin="anonymous"></script>`,
	`<script> 
	dayjs.extend(window.dayjs_plugin_relativeTime);
	dayjs.extend(window.dayjs_plugin_localizedFormat);</script>`,
}

var headTagBootstrap = []string{
	`<script src="https://cdn.jsdelivr.net/npm/jquery@3.6.1/dist/jquery.min.js"
	integrity="sha256-o88AwQnZB+VDvE9tvIXrMQaPlFFSUTR+nldQm1LuPXQ=" crossorigin="anonymous"></script>`,
	`<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.2.2/dist/css/bootstrap.min.css" rel="stylesheet"
	integrity="sha384-Zenh87qX5JnK2Jl0vWa8Ck2rdkQ2Bzep5IDxbcnCeuOxjzrPF/et3URy9Bv1WTRi" crossorigin="anonymous">`,
	`<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.2.2/dist/js/bootstrap.bundle.min.js"
	integrity="sha384-OERcA2EqjJCMA+/3y+gxIOqMEjwtxJY7qPCqsdltbNJuaOe923+mo//f6V8Qbsw3"
	crossorigin="anonymous"></script>`,
}

var headTagCustom = []string{
	`<link rel="stylesheet" href="/style.css">`,
	`<link rel="stylesheet" href="/dashboard.css">`,
}

var staticFiles = []string{
	"/app-worker.js",
	"/app.js",
	"/app.css",
	"/manifest.webmanifest",
	"/wasm_exec.js",
}
