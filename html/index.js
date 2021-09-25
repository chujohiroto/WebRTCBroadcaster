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
            console.log(d)
            await pc.setLocalDescription(d)
            var answer = await postData("/sdp", btoa(JSON.stringify(pc.localDescription)))

            console.log(JSON.parse(atob(answer)))

            try {
                await pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(answer))))
            } catch (e) {
                alert(e)
            }
        })
        .catch(log)

    pc.ontrack = function (event) {
        const el = document.getElementById('video1');
        el.srcObject = event.streams[0]
        el.autoplay = true
        el.controls = true
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

    return response.text(); // レスポンスの JSON を解析
}
