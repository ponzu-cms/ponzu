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


// Returns a local partial time based on unix timestamp (i.e. HH:MM:SS)
function getPartialTime(unix) {
    var date = new Date(unix);
    var parts = [];
    parts.push(date.getHours());
    parts.push(date.getMinutes());
    parts.push(date.getSeconds());

    return parts.join(":");
}

// Returns a local partial date based on unix timestamp (YYYY-MM-DD)
function getPartialDate(unix) {
    var date = new Date(unix);
    var parts = [];
    parts.push(date.getFullYear());
    
    var month = date.getMonth()+1;
    if (month < 10) {
        month = "0" + String(month);
    }
    parts.push(month);

    var day = date.getDate();
    if (day < 10) {
        day = "0" + String(day);
    }
    parts.push(day);

    return parts.join("-");
}