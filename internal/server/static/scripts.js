let root = '';
let pathName = '';
let videoID = '';

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
    fetchFileList("enter", filename).then(r => {});
}

let player = videojs('#video');
player.on('loadedmetadata', function() {
    //player.currentTime(0);
    //player.play();
});

let lastVideoTmReport = 0;

player.on("timeupdate", function(event) {
    if (Math.floor(Date.now() / 1000) - lastVideoTmReport > 2) {
        saveVideoTm(player.currentTime()).then(() => {});
        lastVideoTmReport = Math.floor(Date.now() / 1000);
    }
})
player.on('ended', function() {
    saveVideoTm(0).then(() => {});
    changeVideoSrc('');
});

player.on('error', function() {
    console.log('XXXXX',player.error);
});


videojs.hook('beforeerror', (player, err) => {
    if (err !== null) {
        const errMsg = err.message + err.toString();
        if (errMsg.includes('No compatible source') || errMsg.includes('Format error')
            || errMsg.includes('NotSupportedError')) {
            console.log('---8', "No compatible source", err)
        } else {
            console.log('err--', errMsg)
            console.log('err 2--', err)
        }
    }

    return err
})

function changeVideoSrc(file) {
    player.pause();
    setVideoSrc(root+"/"+file).then(() => {});;
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

function openSMBDialog() {
    document.getElementById('smbDialog').style.display = 'block';
    closeSourceSelectionDialog();
}

function selectOtherSource() {
    alert("其他来源功能尚未实现！");
    closeSourceSelectionDialog();
}

function closeSMBDialog() {
    document.getElementById('smbDialog').style.display = 'none';
}

function testSMBConnection() {
    const address = document.getElementById('smbAddress').value;
    const username = document.getElementById('smbUsername').value;
    const password = document.getElementById('smbPassword').value;

    testRoot({
        rtype: 'smb',
        smb_address: address,
        smb_user: username,
        smb_password: password,
    }).then(()=>{});
}

function confirmSMB() {
    const address = document.getElementById('smbAddress').value;
    const username = document.getElementById('smbUsername').value;
    const password = document.getElementById('smbPassword').value;

    addRoot({
        rtype: 'smb',
        smb_address: address,
        smb_user: user,
        smb_password: password,
    }).then(()=>{
        closeSMBDialog();
    });
}

async function testRoot(data) {
    showLoading();
    try {
        const response = await fetch('/test-root', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data)
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

async function addRoot(data) {
    showLoading();
    try {
        const response = await fetch('/add-root', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            throw new Error('网络响应失败');
        }

        fetchFileList().then(()=>{});
    } catch (error) {
        console.error('获取文件列表失败:', error);
        alert('无法获取文件列表，请稍后重试');
    } finally {
        hideLoading(); // 请求完成（成功或失败）时隐藏遮罩
    }
}

function showDirectoryDialog() {
    closeSourceSelectionDialog();
    document.getElementById('directoryDialog').style.display = 'block';
}

function closeDirectoryDialog() {
    document.getElementById('directoryDialog').style.display = 'none';
}

function testDirectoryName() {
    const directoryName = document.getElementById('directoryName').value;
    console.log("测试目录名称:", directoryName);
    testRoot({
        rtype: 'local',
        local_path: directoryName,
    }).then(()=>{});
}

function confirmDirectoryName() {
    const directoryName = document.getElementById('directoryName').value;

    addRoot({
        rtype: 'local',
        local_path: directoryName,
    }).then(()=>{
        closeDirectoryDialog();
    });
}

window.onload = function() {
    fetchFileList().then(()=>{});
};