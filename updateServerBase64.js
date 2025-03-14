const fs = require('fs');
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
  const base64Regex = /^(?:[A-Za-z0-9+\/]{4})*(?:[A-Za-z0-9+\/]{2}==|[A-Za-z0-9+\/]{3}=)?$/;
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
  for (const url of urls) {
    console.log(`正在获取：${url}`);
    const content = await fetchAndProcess(url);
    allContent += content + "\n";
  }
  // 按行拆分，去除空行及重复行
  const lines = allContent.split('\n').map(line => line.trim()).filter(line => line !== "");
  const uniqueLines = Array.from(new Set(lines));
  const finalContent = uniqueLines.join('\n');

  // 将合并后的内容转换为 Base64 编码
  const finalBase64 = Buffer.from(finalContent, 'utf8').toString('base64');
  fs.writeFileSync('server_base64.txt', finalBase64, 'utf8');
  console.log("server_base64.txt 文件更新成功。");
}

main();
