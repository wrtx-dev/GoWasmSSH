import { render } from 'preact';

import './style.css';
import './xterm-fixes.css';
import { useEffect, useId, useRef, useState } from 'preact/hooks';
import { useXTerm } from './hooks/useTerm';
import DirSvg from './assets/dir.svg';
import SearchImg from "./assets/search.svg";
import Down from "./assets/downfill.svg"
import { useCallback } from 'react';

const defaultProxy = window.location.protocol === "https:" ? "wss://" + window.location.host + "/ws" : "ws://" + window.location.host + "/ws";

export function App() {
	const ref = useRef(null);
	const { term, fit, search } = useXTerm(ref);
	const dialogRef = useRef(null);
	const formRef = useRef(null);
	const clientRef = useRef(null);
	const sftpRef = useRef(null);
	const sftpDialogRef = useRef(null);
	const [connected, setConnected] = useState(false);
	const [files, setFiles] = useState([]);
	const [currentPath, setCurrentPath] = useState(undefined);
	const msgboxRef = useRef(null);
	const [msg, setMsg] = useState(null);
	const confirmRef = useRef(null);
	const [confirmInfo, setConfirmInfo] = useState(null);
	const [connecedSftp, setConnectedSftp] = useState(false);
	const [ignoreCase, setIgnoreCase] = useState(true);
	const [fullWordMatch, setFullWordMatch] = useState(false);
	const [findData, setFindData] = useState("");

	const sftpItemClick = (file) => {
		if (file.isDir && sftpRef && sftpRef.current) {
			const p = `${currentPath}${file.name.startsWith("/") || currentPath === "/" ? "" : "/"}${file.name}`;
			console.log("currentPath:", currentPath, "file.name:", file.name, "p:", p);
			sftpRef.current.list(file.path, (files) => {
				setFiles(files);
				setCurrentPath(p);
			});
		} else {
			if (sftpRef && sftpRef.current) {
				sftpRef.current.download(file.path, (data) => {
					const blob = new Blob([data], { type: 'application/octet-stream' });
					const url = URL.createObjectURL(blob);
					const a = document.createElement('a');
					a.href = url;
					a.download = file.name;
					a.click();
					URL.revokeObjectURL(url);
				});
			}
		}
	}

	const connectHandler = useCallback((e) => {
		e.preventDefault();
		const formData = new FormData(formRef.current);

		let host = formData.get('host');
		if (host && host.toString().includes(':') && !host.toString().startsWith('[') && !host.toString().endsWith(']')) {
			host = `[${host.toString()}]`;
		}
		const port = formData.get('port');
		const username = formData.get('username');
		const password = formData.get('password');
		const proxy = formData.get('proxy');
		const keyFile = formData.get('key');
		const passphrase = formData.get('Passphrase');
		const finger = formData.get("finger");
		const createSftp = formData.get("usesftp");


		dialogRef.current.close();
		// @ts-ignore
		const client = sshNewConnection(proxy);
		clientRef.current = client;
		client.setHostInfo(host.toString(), parseInt(port.toString()));
		client.setUserPassword(username, password);
		client.setTerminal(term.current);
		client.setShowFingerPrint(finger === "on");
		client.setCallback((title, msg) => {
			setMsg({ title, info: msg });
			msgboxRef.current.showModal();
		}, (title, msg, action) => {
			setConfirmInfo({ title, info: msg, action });
			confirmRef.current.showModal();
		}, (flag) => {
			setConnected(flag);
			if (createSftp) {
				setConnectedSftp(flag);
			}
			if (!flag) {
				clientRef.current = null;
				sftpRef.current = null;
			}
		});
		if (createSftp) {
			client.createSftClient();
		}
		if (term && term.current) {
			term.current.onData((data) => {
				client.sessionInput(data);
			});
			term.current.focus();
			term.current.clear();
		}
		(async () => {
			if (keyFile && keyFile instanceof File) {
				const file = keyFile;
				if (file.size > 0) {
					const text = await file.text();
					if (passphrase) {
						client.setPrivateKey(text, passphrase.toString());
					} else {
						client.setPrivateKey(text);
					}
				}
			}
			client.connect();
		})();
	}, []);

	// 清空表单的函数
	const clearForm = () => {
		if (formRef.current) {
			formRef.current.reset();
		}
	};
	const resize = () => {
		if (fit && fit.current) {
			fit.current.fit();
			if (clientRef && clientRef.current) {
				clientRef.current.resize();
			}
		}
	}
	useEffect(() => {
		window.addEventListener("resize", resize);
		return () => {
			window.removeEventListener("resize", resize);
		}
	}, []);
	const menuId = useId();
	return (
		<main className={"h-screen w-screen flex flex-col overflow-hidden border-none"} data-theme={"light"}>
			<nav className={"navbar bg-primary text-primary-content shadow-sm"}>
				<div className={"navbar-start"}>
					<a className={"text-3xl font-mono select-none cursor-default"}>GoWasmSSH</a>
				</div>
				<div className={"navbar-center "}>
					<div
						className={"grow join input rounded-sm text-black dark:text-white border-black focus-within:border-black focus-within:outline-0 focus-within:ring-0 items-center px-1 gap-0"}
					>
						<input
							type="text"
							className={"input join-item border-none focus-within:border-none focus-within:outline-0 focus-within:ring-0 px-1"}
							placeholder={"Search"}
							value={findData}
							onChange={(e) => {
								//@ts-ignore
								setFindData(e.target.value);
							}}

						/>
						<button
							className={"join-item btn btn-success rounded-s-sm btn-sm"}
							disabled={findData === "" || !connected}
							onClick={() => {
								if (search && search.current) {
									search.current.findNext(findData, {
										wholeWord: fullWordMatch,
										caseSensitive: ignoreCase
									})
								}
							}}
						>
							<p className={"select-none"}>查找</p><img src={SearchImg} className={"w-4 h-4"} />
						</button>
						<button
							className={"join-item btn btn-success btn-square rounded-e-sm btn-sm"}
							popoverTarget={menuId}
							style={{ anchorName: "--anchor-1" }}
						>
							<img src={Down} className={"w-4 h-4"} />
						</button>
						<div
							className={"dropdown dropdown-center menu w-52 rounded-box bg-base-100 shadow-sm flex flex-col items-start gap-1.5"}
							popover="auto"
							id={menuId}
							style={{ positionAnchor: "--anchor-1" }}
						>
							<div className={"inline-flex flex-row gap-2 text-xs"}>
								<input
									type="checkbox"
									className={"checkbox checkbox-sm checkbox-neutral"}
									checked={ignoreCase}
									onInput={() => {
										setIgnoreCase(prev => !prev);
									}}
								/>
								<label
									className={"label"}>忽略大小写</label>
							</div>
							<div className={"inline-flex flex-row gap-2 text-xs"}>
								<input
									type="checkbox"
									className={"checkbox checkbox-sm checkbox-neutral"}
									checked={fullWordMatch}
									onInput={() => {
										setFullWordMatch(prev => !prev);
									}} />
								<label className={"label"}>全字匹配</label>
							</div>
						</div>
					</div>
				</div>
				<div className={"navbar-end gap-1"}>
					{!connected ?
						<button
							className={"btn btn-success btn-md focus-within:ring-0 focus-within:outline-0"}
							onClick={() => {
								dialogRef.current.showModal();
							}}
						>
							连接
						</button>
						:
						<button
							className={"btn btn-error btn-md focus-within:ring-0 focus-within:outline-0"}
							onClick={() => {
								if (sftpRef && sftpRef.current) {
									sftpRef.current.close();
								}
								if (clientRef.current) {
									clientRef.current.disconnect();
								}
							}}
						>
							断开
						</button>
					}
					<button
						className={"btn btn-soft btn-info btn-square"}
						disabled={!connected || !connecedSftp}
						onClick={() => {
							if (sftpRef && sftpRef.current) {
								sftpDialogRef.current.showModal();
							}
							if (clientRef && clientRef.current && sftpRef && !sftpRef.current) {
								(async () => {
									const sftp = clientRef.current.sftClient();
									if (sftp) {
										sftpRef.current = sftp;
										sftp.cwd(setCurrentPath);
										sftp.list(".", setFiles);
										sftpDialogRef.current.showModal();
									}
								})();
							}
						}}
					>
						<img src={DirSvg} className={"w-6 h-6"} />
					</button>
				</div>
			</nav>
			<dialog className={"modal pt-1.5"} ref={dialogRef}>

				<div className={"modal-box pb-1 px-2 w-3/5 max-w-3xl pt-1"}>
					<div className={"flex flex-col gap-2 w-full"}>
						<span
							className={"modal-top w-full font-semibold text-xl font-mono pb-2 inline-flex flex-row justify-center items-center cursor-default select-none"}
						>
							设置SSH连接信息
						</span>
					</div>
					<form
						className={"flex flex-col gap-2 w-full text-xs"}
						ref={formRef}
						onSubmit={connectHandler}
					>

						<div className={"grid grid-cols-2 gap-1"}>
							{/*host*/}
							<div className={"col-span-1 join rounded-sm gap-0.5 input border-black focus-within:border-black focus-within:outline-0 focus-within:ring-0 items-center w-full"}>
								<label className={"label join-item"}>主机地址</label>
								<input
									type="text"
									className={"input join-item border-none focus-within:border-none focus-within:outline-0 focus-within:ring-0 px-1"}
									name="host"
								/>
							</div>
							{/*port*/}
							<div className={"col-span-1 join rounded-sm gap-0.5 input border-black focus-within:border-black focus-within:outline-0 focus-within:ring-0 items-center w-full"}>
								<label className={"label join-item"}>端口</label>
								<input
									type="number"
									defaultValue={22}
									className={"input join-item border-none focus-within:border-none focus-within:outline-0 focus-within:ring-0 ps-1"}
									name={"port"}
								/>
							</div>
						</div>


						<div className={"grid grid-cols-2 gap-1"}>
							{/*usename*/}
							<div className={"col-span-1 join rounded-sm gap-0.5 input border-black focus-within:border-black focus-within:outline-0 focus-within:ring-0 items-center w-full"}>
								<label className={"label join-item"}>用户名</label>
								<input
									type="text"
									className={"input join-item border-none focus-within:border-none focus-within:outline-0 focus-within:ring-0 px-1"}
									autoComplete={"username"}
									name={"username"}
								/>
							</div>
							{/*paasword*/}
							<div className={"col-span-1 join rounded-sm gap-0.5 input border-black focus-within:border-black focus-within:outline-0 focus-within:ring-0 items-center w-full"}>
								<label className={"label join-item"}>密码</label>
								<input
									type="password"
									className={"input join-item border-none focus-within:border-none focus-within:outline-0 focus-within:ring-0 px-1"}
									autocomplete="current-password"
									name={"password"}
								/>
							</div>
						</div>

						<div className={"grid grid-cols-2 gap-1"}>
							{/* 私钥*/}
							<div className={"col-span-1 join rounded-sm gap-0.5 input border-black focus-within:border-black focus-within:outline-0 focus-within:ring-0 items-center w-full"}>
								<label className={"label join-item"}>私钥</label>
								<input
									type="file"
									className={"file-input file-input-xs file-input-ghost join-item border-none focus-within:border-none focus-within:outline-0 focus-within:ring-0 px-1"}
									name={"key"}
								/>
							</div>

							{/*Passphrase*/}
							<div className={"col-span-1 join rounded-sm gap-0.5 input border-black focus-within:border-black focus-within:outline-0 focus-within:ring-0 items-center w-full"}>
								<label className={"label join-item"}>密钥口令</label>
								<input
									type="password"
									className={"input join-item border-none focus-within:border-none focus-within:outline-0 focus-within:ring-0 px-1"}
									autocomplete="current-password"
									name={"Passphrase"}
								/>
							</div>

						</div>
						<div className={"grid grid-cols-2 gap-1 items-center"}>
							<div className={"col-span-1  rounded-sm border-black inline-flex flex-row items-center justify-start gap-2"}>
								<input type='checkbox' className={"checkbox "} defaultChecked name="finger" />
								<label className={"label "}>显示ssh指纹信息</label>
							</div>
							<div className={"col-span-1  rounded-sm border-black inline-flex flex-row items-center justify-start gap-2"}>
								<input type='checkbox' className={"checkbox "} name="usesftp" />
								<label className={"label "}>创建sftp连接</label>
							</div>
						</div>

						<div className={"inline-flex flex-row w-full join input border-black focus-within:border-black focus-within:outline-0 focus-within:ring-0 items-center rounded-sm"}>
							<label className={"label join-item cursor-pointer gap-1"}>
								代理地址
							</label>
							<input
								className={"input join-item border-none focus-within:border-none focus-within:outline-0 focus-within:ring-0 px-1"}
								name={"proxy"}
								defaultValue={defaultProxy}
							/>
						</div>

						<div className={"inline-flex flex-row-reverse w-full gap-2"}>
							<button
								className={"btn btn-sm btn-neutral"}
								type="button"
								onClick={(e) => {
									e.preventDefault();
									dialogRef.current.close();
								}}
							>
								取消
							</button>
							<button
								className={"btn btn-sm btn-warning"}
								type="button"
								onClick={(e) => {
									e.preventDefault();
									clearForm();
								}}
							>
								重置
							</button>
							<button className={"btn btn-sm btn-success"} type="submit">
								连接
							</button>
						</div>
					</form>


				</div>
			</dialog>
			{/*sftp dialog*/}
			<dialog className={"modal pt-1.5 overflow-hidden"} ref={sftpDialogRef}>
				<div className={"modal-box pb-1 px-2 w-3/5 max-w-3xl pt-1 min-h-4/5 bg-base-200 max-h-4/5 h-4/5 overflow-hidden"}>
					<div className={`w-full h-full flex flex-col justify-start items-start rounded-md border-l border-base-300 overflow-hidden`}>

						<h3 className="text-lg font-bold mb-4 w-full inline-flex items-center justify-center text-center align-middle">
							sftp文件浏览器
						</h3>
						{currentPath && <p
							className="text-xs"
						>
							当前路径: {currentPath}
						</p>}
						{currentPath && currentPath !== "/" && <div className={"tooltip  tooltip-info"} data-tip="返回上一层">
							<p
								className="hover-bordered"
								onClick={() => {
									const p = getParentPath(currentPath);
									sftpRef.current.list(p, setFiles);
									setCurrentPath(p)
								}}
							>
								<a className="truncate cursor-pointer hover:bg-base-300 select-none">
									⬆️.. (上级目录)
								</a>
							</p>
						</div>}
						<div className={"w-full h-full flex-1 overflow-y-auto p-1 border border-neutral-300 rounded-sm"}>
							<div className="flex flex-col w-full h-full overflow-x-hidden overflow-y-auto rounded-box ps-2 flex-1 gap-2">
								<table className={"table table-fixed table-pin-rows"}>
									<thead>
										<tr>
											<th className={"w-2/5"}>名称</th>
											<th className={"w-1/5"}>类型</th>
											<th className={"w-1/5"}>大小</th>
											<th className={"w-1/5"}>权限</th>
										</tr>
									</thead>
									<tbody>
										{files.map((file, index) => (
											<tr
												className={"w-full hover:bg-base-300 hover:cursor-pointer max-w-full overflow-hidden"}
												key={`${index}-${file.name}`}
												onClick={() => sftpItemClick(file)}
											>
												<td
													className={"whitespace-nowrap text-ellipsis overflow-hidden"}
													title={file.name}
												>
													{file.isDir ? "📁" : "📄"}{file.name}
												</td>
												<td>{file.isDir ? "文件夹" : "文件"}</td>
												<td>{file.size}</td>
												<td>{file.mode}</td>
											</tr>
										))}
									</tbody>
								</table>
							</div>
						</div>
						{/*footer*/}
						<div className={"w-full inline-flex flex-row-reverse gap-2 p-2"}>
							<button
								className={"btn btn-error btn-sm"}
								onClick={() => {
									sftpDialogRef.current.close();
								}}
							>
								关闭
							</button>
							<button
								className={"btn btn-info btn-sm"}
								onClick={() => {
									//Todo 打开文件选择，然后读取文件内容
									const input = document.createElement("input");
									input.type = "file";
									input.accept = "*";
									input.addEventListener("change", async (e) => {
										console.log("in change event:", e);
										//@ts-ignore
										const file = e.target.files[0];
										if (file) {
											const reader = new FileReader();
											reader.onload = (e) => {
												const data = e.target.result;
												if (sftpRef && sftpRef.current) {
													const path = `${currentPath}/${file.name}`;
													console.log("path:", path, "uploade:", sftpRef.current);
													sftpRef.current.upload(path, data, () => {
														sftpRef.current.list(currentPath, setFiles);
													});
												}
											};
											reader.readAsArrayBuffer(file);
										}
									});
									input.click();
								}}
							>
								上传文件
							</button>
						</div>

					</div>
				</div>

			</dialog>


			<div className={"terminal-container bg-black"} ref={ref} />

			<dialog className={"modal"} ref={msgboxRef}>
				<div className={"modal-box"}>
					<h3 className="font-bold text-lg">{msg ? msg.title : ""}</h3>
					<p className="py-4 overflow-x-hidden break-words whitespace-pre-wrap">{msg ? msg.info : ""}</p>
				</div>
				<form method="dialog" className="modal-backdrop">
					<button>close</button>
				</form>
			</dialog>

			<dialog className={"modal"} ref={confirmRef}>
				<div className={"modal-box"}>
					<h3 className="font-bold text-lg">{confirmInfo ? confirmInfo.title : ""}</h3>
					<p className="py-4 overflow-x-hidden break-words whitespace-pre-wrap">{confirmInfo ? confirmInfo.info : ""}</p>
					<div className={"w-full inline-flex flex-row-reverse gap-2 p-2"}>
						<button
							className={"btn btn-success btn-sm"}
							onClick={() => {
								if (confirmInfo && confirmInfo.action) {
									confirmInfo.action(true);
								}
								confirmRef.current.close();
							}}
						>
							接受
						</button>
						<button
							className={"btn btn-error btn-sm"}
							onClick={() => {
								if (confirmInfo && confirmInfo.action) {
									confirmInfo.action(false);
								}
								confirmRef.current.close();
							}}
						>
							拒绝
						</button>
					</div>
				</div>

			</dialog>

		</main>
	);
}

render(<App />, document.getElementById('app'));

function getParentPath(path) {
	if (typeof path !== 'string') {
		throw new Error("Input must be a string.");
	}
	let normalizedPath = path.replace(/\/+/g, '/');

	if (normalizedPath === '/') {
		return '/';
	}
	if (normalizedPath.endsWith('/') && normalizedPath.length > 1) {
		normalizedPath = normalizedPath.slice(0, -1);
	}
	const lastSlashIndex = normalizedPath.lastIndexOf('/');
	if (lastSlashIndex === -1) {
		return '';
	} else if (lastSlashIndex === 0) {
		return '/';
	} else {

		return normalizedPath.substring(0, lastSlashIndex);
	}
}
