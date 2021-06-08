var lastVersion;
function trimPrefix(str, prefix) {
	if (str.startsWith(prefix)) {
		return str.slice(prefix.length)
	} else {
		return str
	}
}

function noteID() {
	return encodeURIComponent(trimPrefix(location.pathname, '/notes/'));
}

function getNote() {
	$.get('/note?n=' + noteID())
		.then(function (data) {
			document.getElementById('textnote').value = data;
		})
		.fail(HandleFail);
}


function doSendData(callback) {
	$.post('/note?n=' + noteID(), document.getElementById('textnote').value)
		.then(data => lastVersion = new Date(data))
		.fail(HandleFail);
}
var pendingClick;
function sendDataDelayed() {
	clearTimeout(pendingClick);
	pendingClick = setTimeout(doSendData, 1000);
}


checkVersion();
var intervalId = window.setInterval(checkVersion, 5000);
function checkVersion() {
	$.get('/note/v?n=' + noteID())
		.then(function (timestamp) {
			let date = new Date(timestamp);
			if (lastVersion === undefined || (timestamp != '' && date > lastVersion)) {
				lastVersion = date
				getNote();
			}
		})
		.fail(HandleFail);
}

function HandleFail(error) {
	let err = error.responseJSON ? error.responseJSON : "Oops! unkown error.";
	$('#error').show(400).text("Error: " + err);
	setTimeout(() => { $('#error').hide(400) }, 3000);
}