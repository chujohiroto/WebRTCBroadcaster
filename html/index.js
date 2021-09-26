/* eslint-env browser */
const pc = new RTCPeerConnection({
    iceServers: [
        {
            urls: 'stun:stun.l.google.com:19302'
        }
    ]
})

pc.onconnectionstatechange = event => {
    console.log(pc.connectionState)
}

pc.oniceconnectionstatechange = _ => console.log(pc.iceConnectionState)

pc.onicecandidate = async event => {
    if (event.candidate === null) {
        await window.startSession()
    }
}

pc.ontrack = function (event) {
    console.log(event)
    const el = document.createElement(event.track.kind)
    el.srcObject = event.streams[0]
    el.autoplay = true
    el.controls = true
    document.getElementById('remoteVideos').appendChild(el)
}

pc.addTransceiver('video', {
    'direction': 'recvonly'
})

pc.createOffer()
    .then(async d => {
        console.log(d)
        await pc.setLocalDescription(d)
    })
    .catch(console.log)

window.startSession = async () => {
    const answer = await postData("/sdp",
        {
            "sdp_offer": btoa(JSON.stringify(pc.localDescription)),
            "authnMetadata" : {
                "user": "example"
            }
        })
    console.log(new RTCSessionDescription(JSON.parse(atob(answer))))

    try {
        await pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(answer))))
    } catch (e) {
        alert(e)
    }
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
        },
        redirect: 'follow', // manual, *follow, error
        referrerPolicy: 'no-referrer', // no-referrer, *no-referrer-when-downgrade, origin, origin-when-cross-origin, same-origin, strict-origin, strict-origin-when-cross-origin, unsafe-url
        body: JSON.stringify(data) // 本文のデータ型は "Content-Type" ヘッダーと一致する必要があります
    })

    return response.text(); // レスポンスの JSON を解析
}
