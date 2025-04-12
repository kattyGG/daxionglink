const fs = require('fs');
const path = require('path');
const fetch = require('node-fetch');

// 要获取的订阅文件列表
const urls = [
  "https://raw.githubusercontent.com/mfuu/v2ray/master/v2ray",
  "https://raw.githubusercontent.com/peasoft/NoMoreWalls/master/list_raw.txt",
  "https://raw.githubusercontent.com/ermaozi/get_subscribe/main/subscribe/v2ray.txt",
  "https://raw.githubusercontent.com/aiboboxx/v2rayfree/main/v2",
  "https://raw.githubusercontent.com/mahdibland/SSAggregator/master/sub/airport_sub_merge.txt",
  "https://raw.githubusercontent.com/mahdibland/SSAggregator/master/sub/sub_merge.txt",
  "https://raw.githubusercontent.com/Pawdroid/Free-servers/refs/heads/main/sub"
];

// 判断字符串是否为有效的 base64 编码
function isBase64(str) {
  const s = str.trim().replace(/\s/g, '');
  // 简易正则判断
  const base64Regex = /^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$/;
  return base64Regex.test(s);
}

// 获取指定 URL 的内容，并检测是否需要 base64 解码
async function fetchAndProcess(url) {
  try {
    const response = await fetch(url);
    if (!response.ok) {
      console.error(`获取 ${url} 失败: ${response.statusText}`);
      return "";
    }
    let text = await response.text();
    // 如果内容是 base64 格式则进行解码
    if (isBase64(text)) {
      try {
        text = Buffer.from(text, 'base64').toString('utf8');
      } catch (err) {
        console.error(`解码 ${url} 的 base64 内容出错: ${err}`);
      }
    }
    return text;
  } catch (err) {
    console.error(`请求 ${url} 时发生错误: ${err}`);
    return "";
  }
}

async function main() {
  let allContent = "";
  // 1. 逐个获取并合并内容
  for (const url of urls) {
    console.log(`正在获取：${url}`);
    const content = await fetchAndProcess(url);
    allContent += content + "\n";
  }

  // 2. 按行拆分，去除空行及重复行
  const lines = allContent
    .split('\n')
    .map(line => line.trim())
    .filter(line => line !== "");
  const uniqueLines = Array.from(new Set(lines));

  // 3. 将合并后的内容转换为 Base64 并写入大文件 server_base64.txt
  const finalContent = uniqueLines.join('\n');
  const finalBase64 = Buffer.from(finalContent, 'utf8').toString('base64');
  fs.writeFileSync('server_base64.txt', finalBase64, 'utf8');
  console.log("server_base64.txt 文件更新成功。");

  // 4. 再将最终内容按每 500 行切分成多个文件
  //    并将每个小文件的内容再 Base64 编码，写到 serverlist/server_base64_XX.txt
  const chunkSize = 500;
  // 如果不存在 serverlist 目录，就创建
  if (!fs.existsSync('serverlist')) {
    fs.mkdirSync('serverlist');
  }

  for (let i = 0; i < uniqueLines.length; i += chunkSize) {
    // 取出 500 行
    const chunkLines = uniqueLines.slice(i, i + chunkSize);
    if (chunkLines.length === 0) break;

    // 拼成字符串后再进行 Base64
    const chunkContent = chunkLines.join('\n');
    const chunkBase64 = Buffer.from(chunkContent, 'utf8').toString('base64');

    // 生成带序号的文件名，比如 server_base64_01.txt、server_base64_02.txt
    const chunkIndex = Math.floor(i / chunkSize) + 1;
    const fileName = `server_base64_${String(chunkIndex).padStart(2, '0')}.txt`;

    // 写入 serverlist/ 目录下
    const filePath = path.join('serverlist', fileName);
    fs.writeFileSync(filePath, chunkBase64, 'utf8');

    console.log(`已生成：${filePath}`);
  }

  console.log("小文件全部生成完毕。");

  //zip文件
  for (let i = 0; i < uniqueLines.length; i += chunkSize) {
    const chunkLines = uniqueLines.slice(i, i + chunkSize);
    if (chunkLines.length === 0) break;

    // 直接拼接成字符串，不进行 Base64 编码
    const chunkContent = chunkLines.join('\n');

    // 生成带序号的文件名，比如 server_original_01.txt、server_original_02.txt
    const chunkIndex = Math.floor(i / chunkSize) + 1;
    const fileName = `server_original_${String(chunkIndex).padStart(2, '0')}.txt`;
    const filePath = path.join('serverlist', fileName);
    fs.writeFileSync(filePath, chunkContent, 'utf8');
    console.log(`已生成：${filePath}`);
  }

  console.log("小文件全部生成完毕。");

  // 5. 将 serverlist 目录下的所有小文件压缩成一个 zip 文件，并设置密码为 "daxionglink"
  const zipPath = path.join(__dirname, 'serverlist.zip');
  const output = fs.createWriteStream(zipPath);
  const archive = archiver('zip-encrypted', {
    zlib: { level: 9 },
    encryptionMethod: 'aes256',
    password: 'daxionglink'
  });

  output.on('close', () => {
    console.log(`压缩文件创建成功，总大小：${archive.pointer()} 字节`);
  });

  archive.on('error', (err) => {
    throw err;
  });

  archive.pipe(output);
  // 将 serverlist 目录下的所有文件加入压缩包（不包含目录结构）
  archive.directory('serverlist/', false);
  await archive.finalize();
}

main();
