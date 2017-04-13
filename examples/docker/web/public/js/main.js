(function() {

    function loadHomepage() {
        const xhr = new XMLHttpRequest();
        xhr.open('GET', '/api/contents?type=Song');
        xhr.onreadystatechange = renderHomepage;
        xhr.send(null);
    }

    function renderHomepage(event) {
        const DONE = 4;
        const OK = 200;
        let xhr = event.currentTarget;
        let html = '';

        if (xhr.readyState === DONE) {
            if (xhr.status === OK) {
                const songs = window.JSON.parse(xhr.responseText).data;
                console.log(songs);
                if(songs.length === 0){
                    html = '<p><strong>There have not been any Songs added, <a href="/admin">go add some at /admin</a></strong></p>';
                } else {
                    html = songs.map(function(song) {
                        return `
                            <article>
                                <h3>${song.title || 'Unknown'} by ${song.artist || 'Unknown'}</h3>
                                <p>rating: ${song.rating}</p>
                                <h6>opinion:</h6>
                                <div>${song.opinion || 'none'}
                            </article>
                        `;
                    }).join();
                }
            } else {
                html = '<p><strong>The /api endpoint did not respond correctly :-(</strong></p>';
            }

            document.querySelector('#main').innerHTML = html;
        }
    }

    document.addEventListener("DOMContentLoaded", loadHomepage);
})();