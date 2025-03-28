let root = '';
let pathName = '';
let videoID = '';

// 显示/隐藏遮罩的函数
function showLoading() {
    document.getElementById('loadingOverlay').style.display = 'flex';
}

function hideLoading() {
    document.getElementById('loadingOverlay').style.display = 'none';
}

// 创建文件项的HTML
function createFileItem(file) {
    if (file.is_dir) {
        return `
            <li class="file-item" onclick="enterDir('${file.path}')">
                <span class="file-icon">📂</span>
                <span class="file-name">${file.name}</span>
            </li>
        `;
    }

    return `
        <li class="file-item" onclick="openFile('${file.path}')">
            <span class="file-icon">📄</span>
            <span class="file-name">${file.name}</span>
            <span class="file-size">${file.size}</span>
        </li>
    `;
}

async function setVideoSrc(videoURL) {
    showLoading();
    try {
        const response = await fetch('/s-video-id', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                video_url: videoURL,
            })
        });

        if (!response.ok) {
            throw new Error('网络响应失败');
        }

        const resp = await response.json();
        videoID = resp["video_id"];

        player.src({src: '/videos?file=' + videoID,type: 'video/webm'});

        player.load();
        player.currentTime(resp["last_tm"]);
        player.play();
    } catch (error) {
        console.error('获取文件列表失败:', error);
        alert('无法获取文件列表，请稍后重试');
    } finally {
        hideLoading(); // 请求完成（成功或失败）时隐藏遮罩
    }
}

async function saveVideoTm(tm) {
    try {
        await fetch('/video/save-tm', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                vid: videoID,
                tm:  Math.floor(tm),
            })
        });
    } catch (error) {
        console.error('save video tm failed:',videoID, tm, error);
    } finally {
    }
}

// 通过HTTP获取最新文件列表
async function fetchFileList(op, dir) {
    showLoading(); // 请求开始时显示遮罩
    try {
        const response = await fetch('/browser', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                op: op,
                path: root,
                dir: dir
            })
        });

        if (!response.ok) {
            throw new Error('网络响应失败');
        }

        const resp = await response.json();
        root = resp["path"];
        pathName = resp["pathName"]
        const files = resp["items"];

        // 更新列表
        const rootUI = document.getElementById('root');
        rootUI.innerHTML = pathName;

        const fileList = document.getElementById('fileList');
        fileList.innerHTML = '';
        files.forEach(file => {
            fileList.innerHTML += createFileItem(file);
        });
    } catch (error) {
        console.error('获取文件列表失败:', error);
        alert('无法获取文件列表，请稍后重试');
    } finally {
        hideLoading(); // 请求完成（成功或失败）时隐藏遮罩
    }
}

function enterDir(filename) {
    fetchFileList("enter", filename);
}

let player = videojs('#video');
player.on('loadedmetadata', function() {
    //player.currentTime(0);
    //player.play();
    console.log('---4')
});

let lastVideoTmReport = 0;
player.on("timeupdate", function(event) {
    console.log(player.currentTime())
    if (Math.floor(Date.now() / 1000) - lastVideoTmReport > 2) {
        saveVideoTm(player.currentTime())
        lastVideoTmReport = Math.floor(Date.now() / 1000);
    }
})
player.on('ended', function() {
    saveVideoTm(0);
    console.log('Awww...over so soon?!');
    changeVideoSrc('');
});

player.on('error', function(event) {
    console.log('-------5', event);
});

videojs.hook('beforeerror', (player, err) => {
    console.log('hook - beforeerror', player.src(), err)
    // Video.js 在切换/指定 source 后立即会触发一个 err=null 的错误，这里过滤一下
    if (err !== null) {
        if (err.message.includes('No compatible source')) {
            console.log('---8', "No compatible source")
        }
        //player.src(sources[++index])
    }

    // 清除错误，避免 error 事件在控制台抛出错误
    return null
})

function changeVideoSrc(file) {
    player.pause();
    setVideoSrc(root+"/"+file);
}

function openFile(filename) {
    document.getElementById('file-list-container').style.display = 'none';
    document.getElementById('video-container').style.display = 'block';
    changeVideoSrc(filename);
}

function videoBack() {
    player.pause();
    document.getElementById('video-container').style.display = 'none';
    document.getElementById('file-list-container').style.display = 'block';
}

function openSourceSelectionDialog() {
    document.getElementById('sourceSelectionDialog').style.display = 'block';
}

function closeSourceSelectionDialog() {
    document.getElementById('sourceSelectionDialog').style.display = 'none';
}

function selectLocalSource() {
    document.getElementById('localDirectory').click();
    closeSourceSelectionDialog();
}

function openSMBDialog() {
    document.getElementById('smbDialog').style.display = 'block';
    closeSourceSelectionDialog();
}

function selectOtherSource() {
    alert("其他来源功能尚未实现！");
    closeSourceSelectionDialog();
}

function showSourceSelection() {
    const choice = confirm("选择来源：点击确定选择本地，点击取消选择 SMB");
    if (choice) {
        document.getElementById('localDirectory').click();
    } else {
        document.getElementById('smbDialog').style.display = 'block';
    }
}

function handleLocalDirectory(event) {
    const files = event.target.files;
    console.log("选择的本地目录文件:", files);
    alert("本地目录选择成功！");
}

function closeSMBDialog() {
    document.getElementById('smbDialog').style.display = 'none';
}

function testSMBConnection() {
    const address = document.getElementById('smbAddress').value;
    const username = document.getElementById('smbUsername').value;
    const password = document.getElementById('smbPassword').value;

    console.log("测试 SMB 连接:", { address, username, password });
    testSMBRoot(address, username, password)
}

function confirmSMB() {
    const address = document.getElementById('smbAddress').value;
    const username = document.getElementById('smbUsername').value;
    const password = document.getElementById('smbPassword').value;

    console.log("确认 SMB 配置:", { address, username, password });
    addSMBRoot(address, username, password).then(()=>{
        closeSMBDialog();
    });
}

async function testSMBRoot(address, user, password) {
    showLoading();
    try {
        const response = await fetch('/test-root', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                rtype: 'smb',
                smb_address: address,
                smb_user: user,
                smb_password: password,
            })
        });

        if (!response.ok) {
            throw new Error('网络响应失败');
        }


        const resp = await response.json();
        const statusCode = resp["status_code"];
        const message = resp["message"];
        if (statusCode === 0) {
            alert('测试成功');

            return;
        }

        alert('测试失败:'+message);
    } catch (error) {
        console.error('获取文件列表失败:', error);
        alert('无法获取文件列表，请稍后重试');
    } finally {
        hideLoading(); // 请求完成（成功或失败）时隐藏遮罩
    }
}

async function addSMBRoot(address, user, password) {
    showLoading();
    try {
        const response = await fetch('/add-root', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                rtype: 'smb',
                smb_address: address,
                smb_user: user,
                smb_password: password,
            })
        });

        if (!response.ok) {
            throw new Error('网络响应失败');
        }

        fetchFileList();
    } catch (error) {
        console.error('获取文件列表失败:', error);
        alert('无法获取文件列表，请稍后重试');
    } finally {
        hideLoading(); // 请求完成（成功或失败）时隐藏遮罩
    }
}

window.onload = function() {
    fetchFileList();
};