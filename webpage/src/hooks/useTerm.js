import "@xterm/xterm/css/xterm.css";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { WebglAddon } from "@xterm/addon-webgl";
import { SearchAddon } from "@xterm/addon-search";
import { WebLinksAddon } from "@xterm/addon-web-links"
import { useRef, useEffect } from "preact/hooks";


const info = `
这是一个基于WebAssembly (WASM) 技术开发的网页版SSH客户端。
您可以使用它通过SSH协议，远程登录到您本地网络之外的任何机器。
由于所有的SSH协议操作都在您的浏览器中进行，服务器无法获取除您的IP和目标服务器的IP之外的任何信息。
服务器仅充当中继角色，将WebSocket流量转换为TCP。
要开始使用，请点击右上角的连接图标。
`
export const useXTerm = (ref) => {
    const termRef = useRef(null);
    const fitRef = useRef(null);
    const searchRef = useRef(null);
    useEffect(() => {
        const terminal = new Terminal({
            cursorBlink: true,
            theme: {
                background: '#000000',
            },
        });
        terminal.open(ref.current);
        const fitAddon = new FitAddon();
        terminal.loadAddon(fitAddon);
        const webgl = new WebglAddon();
        terminal.loadAddon(webgl);
        const linkAddon = new WebLinksAddon();
        terminal.loadAddon(linkAddon);
        const searchAddon = new SearchAddon();
        terminal.loadAddon(searchAddon);
        searchRef.current = searchAddon;

        fitRef.current = fitAddon;
        termRef.current = terminal;
        info.split("\n").forEach(line => {
            terminal.writeln(line);
        })
        fitAddon.fit();
    }, []);
    return {
        term: termRef,
        fit: fitRef,
        search: searchRef,
    };

}