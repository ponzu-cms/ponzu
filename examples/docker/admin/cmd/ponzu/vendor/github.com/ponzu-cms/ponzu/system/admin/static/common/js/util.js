// Replaces commonly-used Windows 1252 encoded chars that do not exist in ASCII or ISO-8859-1 with ISO-8859-1 cognates.
// modified from http://www.andornot.com/blog/post/Replace-MS-Word-special-characters-in-javascript-and-C.aspx
function replaceBadChars(text) {
    var s = text;
    // smart single quotes and apostrophe
    s = s.replace(/[\u2018\u2019\u201A]/g, "\'");
    // smart double quotes
    s = s.replace(/[\u201C\u201D\u201E]/g, "\"");
    // ellipsis
    s = s.replace(/\u2026/g, "...");
    // dashes
    s = s.replace(/[\u2013\u2014]/g, "-");
    // circumflex
    s = s.replace(/\u02C6/g, "^");
    // open angle bracket
    s = s.replace(/\u2039/g, "<");
    // close angle bracket
    s = s.replace(/\u203A/g, ">");
    // spaces
    s = s.replace(/[\u02DC\u00A0]/g, " ");
    
    return s;
}


// Returns a local partial time object based on unix timestamp
function getPartialTime(unix) {
    var date = new Date(unix);
    var t = {};
    var hours = date.getHours();
    if (hours < 10) {
        hours = "0" + String(hours);
    }

    t.hh = hours;
    if (hours > 12) {
        t.hh = hours - 12;
        t.pd = "PM";
    } else if (hours === 12) {
        t.pd = "PM";
    } else if (hours < 12) {
        t.pd = "AM";
    }

    var minutes = date.getMinutes();
    if (minutes < 10) {
        minutes = "0" + String(minutes);
    }
    t.mm = minutes;

    return t;
}

// Returns a local partial date object based on unix timestamp
function getPartialDate(unix) {
    var date = new Date(unix);
    var d = {};
    
    d.yyyy = date.getFullYear();
    
    d.mm = date.getMonth()+1;

    var day = date.getDate();
    if (day < 10) {
        day = "0" + String(day);
    }
    d.dd = day;

    return d;
}

// Returns a part of the window URL 'search' string
function getParam(param) {
    var qs = window.location.search.substring(1);
    var qp = qs.split('&');
    var t = '';

    for (var i = 0; i < qp.length; i++) {
        var p = qp[i].split('=')
        if (p[0] === param) {
            t = p[1];	
        }
    }

    return t;
}