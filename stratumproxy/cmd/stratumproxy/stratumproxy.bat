:while
cd /d %~dp0
stratumproxy.exe --server-pem server.pem --server-key server.key --listen-addr 127.0.0.1:4420 --server-addr 123.123.123.123:4420 --replaced-user " minejs.aleo02" --replaced-password "aleo02"
rem stratumproxy.exe --server-pem server.pem --server-key server.key --listen-addr 127.0.0.1:1177 --server-addr 192.168.8.115:1177 --replaced-user "pyrin:qq0240xcnlk52jt4t007gwe97hnr33g5knx9kkgarmm0p9ghm9sg68qrakyf2.pyi02" --replaced-password "pyi02"
goto while