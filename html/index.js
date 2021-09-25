/* eslint-env browser */
var log = msg => {
    document.getElementById('logs').innerHTML += msg + '<br>'
}

window.createSession = () => {
    let pc = new RTCPeerConnection({
        iceServers: [
            {
                urls: 'stun:stun.l.google.com:19302'
            }
        ]
    })
    pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
    pc.onicecandidate = event => {
        if (event.candidate === null) {
            document.getElementById('localSessionDescription').value = btoa(JSON.stringify(pc.localDescription))
        }
    }

    pc.addTransceiver('video')
    pc.createOffer()
        .then(async d => {
            pc.setLocalDescription(d)
            await postData("http://localhost:8888/sdp", d)
        })
        .catch(log)

    pc.ontrack = function (event) {
        var el = document.getElementById('video1')
        el.srcObject = event.streams[0]
        el.autoplay = true
        el.controls = true
    }

    window.startSession = () => {
        let sd = document.getElementById('remoteSessionDescription').value
        if (sd === '') {
            return alert('Session Description must not be empty')
        }

        try {
            pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(sd))))
        } catch (e) {
            alert(e)
        }
    }

    let btns = document.getElementsByClassName('createSessionButton')
    for (let i = 0; i < btns.length; i++) {
        btns[i].style = 'display: none'
    }

    document.getElementById('signalingContainer').style = 'display: block'
}

async function postData(url = '', data = {}) {
    // 既定のオプションには * が付いています
    const response = await fetch(url, {
        method: 'POST', // *GET, POST, PUT, DELETE, etc.
        mode: 'cors', // no-cors, *cors, same-origin
        cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
        credentials: 'same-origin', // include, *same-origin, omit
        headers: {
            'Content-Type': 'application/json'
            // 'Content-Type': 'application/x-www-form-urlencoded',
        },
        redirect: 'follow', // manual, *follow, error
        referrerPolicy: 'no-referrer', // no-referrer, *no-referrer-when-downgrade, origin, origin-when-cross-origin, same-origin, strict-origin, strict-origin-when-cross-origin, unsafe-url
        body: JSON.stringify(data) // 本文のデータ型は "Content-Type" ヘッダーと一致する必要があります
    })
    return response.json(); // レスポンスの JSON を解析
}
