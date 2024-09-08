import * as ProtoBuf from "protobufjs";
import * as ByteBuffer from "bytebuffer";
const root = ProtoBuf.loadSync("./test.proto");

class packet {
    protected message: any;
    protected pack: any;

    public decode(data: any): [string, any] {
        // 解包
        const bufferArray: Uint8Array = new Uint8Array(data);
        const buffer: ByteBuffer = new ByteBuffer().append(bufferArray);
        const headLen = buffer.readShort(0);
        const cmd = buffer.readString(headLen, 2, 0);
        let resp: any = root.lookupType(`test.${cmd.string}`);
        return (
            cmd.string,
            resp.decode(
                buffer.buffer.slice(2 + cmd.string.length, bufferArray.length)
            )
        );
    }

    public encode(cmd: string, data: any): object {
        let req = root.lookupType(`test.${cmd}`);
        // 设置协议内容
        let pack = req.create(data);
        let buff = req.encode(pack).finish();
        // 打包
        let buffer = new ByteBuffer(cmd.length + 2 + buff.length);
        buffer.writeShort(cmd.length);
        buffer.writeString(cmd);
        buffer.append(buff);
        return buffer.buffer;
    }
}

const Packet = new packet();
export default Packet;
