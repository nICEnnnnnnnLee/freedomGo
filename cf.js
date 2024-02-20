import { connect } from 'cloudflare:sockets';

const WS_READY_STATE_OPEN = 1;
const WS_READY_STATE_CLOSING = 2;
let EXPECTED_USENAME, EXPECTED_PWD, EXPECTED_SALT;

export default {

    async fetch(req, env, _ctx) {
        EXPECTED_USENAME = env["EXPECTED_USENAME"] || "user";
        EXPECTED_PWD = env["EXPECTED_PWD"] || "pwd";
        EXPECTED_SALT = env["EXPECTED_SALT"] || "salt";
        // 舍弃不是WS的消息
        const upgradeHeader = req.headers.get("Upgrade")
        if (upgradeHeader !== "websocket") {
            return new Response("Unexpected connection", { status: 400 })
        }
        // 获取cookie
        const cookieRaw = req.headers.get("Cookie") || "";
        const cookie = {}
        cookieRaw.split(";").forEach(aEqB => {
            const keyVal = aEqB.split("=", 2);
            cookie[keyVal[0].trim()] = keyVal[1].trim();
        });
        // 鉴权
        if (!await isValid(cookie)) {
            return new Response("Invalid connection", { status: 403 })
        }
        // 构造ws并回复
        try {
            const remoteAddr = { hostname: cookie["my_domain"], port: Number(cookie["my_port"]) };
            const socket = connect(remoteAddr);
            const [client, server] = Object.values(new WebSocketPair())
            wsToSocket(server, socket)
            socketToWs(socket, server)

            return new Response(null, {
                status: 101,
                webSocket: client,
                headers: {
                    "auth": "ok",
                }
            })
        } catch (error) {
            return new Response("Socket connection failed: " + error, { status: 500 });
        }
    },
};

async function isValid(cookie) {
    const username = cookie["my_username"] || "", token = cookie["my_token"] || "", timeStr = cookie["my_time"] || "0";
    // 用户名
    if (username !== EXPECTED_USENAME)
        return false;
    // 时间
    const currentTime = Date.now();
    const reqTime = Number(timeStr);
    const deltaTime = currentTime - reqTime;
    if (deltaTime > 600000 || deltaTime < -600000)
        return false
    // token
    const data = new TextEncoder().encode(EXPECTED_PWD + EXPECTED_SALT + timeStr);
    const digest = await crypto.subtle.digest({
        name: 'MD5',
    }, data);
    const EXPECTED_TOKEN = Array.from(new Uint8Array(digest)).map(a=>a.toString(16).padStart(2, "0")).join("");
    if (token !== EXPECTED_TOKEN)
        return false;
    return true;
}

function wsToSocket(webSocket, socket) {
    webSocket.accept();
    const stream = new ReadableStream({
        start(controller) {
            webSocket.addEventListener('message', (event) => {
                controller.enqueue(event.data);
            });
            webSocket.addEventListener('close', () => {
                webSocket.close();
                controller.close();
            });
            webSocket.addEventListener('error', (err) => {
                webSocket.close();
                controller.error(err);
            });
        },
        pull(controller) {
        },
        cancel(reason) {
        }
    });
    stream.pipeTo(new WritableStream({
        async write(chunk, controller) {
            const writer = socket.writable.getWriter();
            await writer.write(chunk);
            writer.releaseLock();
        },
        close() {
        },
        abort(reason) {
        },
    })).catch((err) => {
    });
}

function socketToWs(socket, webSocket) {
    socket.readable
        .pipeTo(
            new WritableStream({
                start() {
                },
                async write(chunk, controller) {
                    if (webSocket.readyState !== WS_READY_STATE_OPEN) {
                        controller.error(
                            'webSocket.readyState is not open, maybe close'
                        );
                    }
                    webSocket.send(chunk);
                },
                close() {
                },
                abort(reason) {
                },
            })
        )
        .catch((error) => {
            console.error(
                `remoteSocketToWS has exception `,
                error.stack || error
            );
            webSocket.close();
        });
}