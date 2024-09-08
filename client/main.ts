import Packet from "./packet";
import * as WebSocket from "ws";
const connection = new WebSocket("ws://127.0.0.1:9501");

connection.onopen = async function () {
    // 登录
    send("C2S_Login", {
        username: "pp",
        password: "pass",
    });
    console.log("send done");
};

connection.onmessage = (event: any) => {
    const resp = event.data;
    if (typeof resp == "string") {
        console.log(`event: ${event} data: ${resp}`);
        return;
    }
    try {
        let cmd,
            data = Packet.decode(resp);
        // 53位长整数问题
        data = fixLongNumber(data);
        console.log(`message: ${cmd} resp: ${JSON.stringify(data)}`);
    } catch (error) {
        console.log(error);
    }
};

function send(cmd: string, data: any) {
    if (typeof cmd == "string") {
        const buffer = Packet.encode(cmd, data);
        connection.send(buffer);
    } else {
        console.log("event is not a string");
    }
}

// 53位
function fixLongNumber(data: any): any {
    let type = typeof data;
    if (type === "object") {
        for (let key in data) {
            let value = data[key];
            let type = value.constructor.name;
            if (type === "object") {
                data[key] = fixLongNumber(value);
                continue;
            }
            if (type === "array") {
                let index = 0;
                for (let item of value) {
                    value[index] = fixLongNumber(item);
                    index++;
                }
            }
            if (type === "Long") {
                let temp = value.toNumber();
                if (!Number.isSafeInteger(temp)) {
                    temp = Number.MAX_SAFE_INTEGER;
                }
                data[key] = temp;
                continue;
            }
        }
    }
    return data;
}

function delay(ms: number) {
    return new Promise((resolve) => setTimeout(resolve, ms));
}
