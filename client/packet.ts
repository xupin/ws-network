import * as ProtoBuf from "protobufjs";
import * as ByteBuffer from "bytebuffer"
const root = ProtoBuf.loadSync("./test.proto");

export default class Packet {

    public protocol: string;
    protected message: any;
    protected pack: any;

    constructor() {
        this.protocol = "";
    }

    public setProtocol(protocol: string): void {
        this.protocol = protocol;
        this.message = root.lookupType(`test.${this.protocol}`);
    }

    public getProtocol(): string {
        return this.protocol;
    }

    public decode(data: any): any {
        // 解包
        const bufferArray: Uint8Array = new Uint8Array(data);
        const buffer: ByteBuffer = new ByteBuffer().append(bufferArray);
        const headLen = buffer.readShort(0);
        const protocol = buffer.readString(headLen, 2, 0);
        // 读取协议内容
        this.setProtocol(protocol.string);
        return this.message.decode(buffer.buffer.slice(2 + protocol.string.length, bufferArray.length));
    }

    public encode(data: any): object {
        // 设置协议内容
        let pack = this.message.create(data);
        let buff = this.message.encode(pack).finish();
        // 打包
        let buffer = new ByteBuffer(this.protocol.length + 2 + buff.length);
        buffer.writeShort(this.protocol.length);
        buffer.writeString(this.protocol);
        buffer.append(buff);
        return buffer.buffer;
    }
}