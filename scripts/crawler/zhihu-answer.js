import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";
import puppeteer from "puppeteer-extra";
import StealthPlugin from "puppeteer-extra-plugin-stealth";
import { withFileLock } from "./utils/withFileLock.js";
import os from "os";

/**
* 1. 安装 puppeteer-extra + stealth 插件
* pnpm install puppeteer  puppeteer-extra puppeteer-extra-plugin-stealth
*
* 2. 安装 chromium-browser
* apt install -y chromium-browser
*/

// 使用 stealth 插件绕过反爬虫检测
puppeteer.use(StealthPlugin({
  enabledEvasions: new Set([
    'chrome.app',
    'chrome.csi',
    'chrome.loadTimes',
    'chrome.runtime',
    'iframe.contentWindow',
    'media.codecs',
    'navigator.hardwareConcurrency',
    'navigator.languages',
    'navigator.permissions',
    'navigator.plugins',
    'navigator.vendor',
    'navigator.webdriver',
    'user-agent-override',
    'webgl.vendor',
    'window.outerdimensions'
  ])
}));

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

/**
 * 延迟函数（替代已弃用的 page.waitForTimeout）
 * @param {number} ms - 延迟毫秒数
 * @returns {Promise<void>}
 */
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * 替换文件名中 Windows 不支持的字符
 * @param {string} filename - 原始文件名
 * @returns {string} 处理后的文件名
 */
function sanitizeFilename(filename) {
  return filename
    .replace(/\//g, "／")  // 斜杠替换为全角斜杠
    .replace(/\\/g, "＼")  // 反斜杠替换为全角反斜杠
    .replace(/\?/g, "？")  // 问号替换为全角问号
    .replace(/:/g, "：")   // 冒号替换为全角冒号
    .replace(/\*/g, "＊")  // 星号替换为全角星号
    .replace(/"/g, "＂")   // 双引号替换为全角双引号
    .replace(/</g, "＜")   // 小于号替换为全角小于号
    .replace(/>/g, "＞")   // 大于号替换为全角大于号
    .replace(/\|/g, "｜")  // 竖线替换为全角竖线
    .trim();
}

/**
 * 使用 Puppeteer 获取页面内容
 * @param {string} url - 要爬取的 URL
 * @param {boolean} debug - 是否开启调试模式
 * @returns {Promise<Object>} 返回解析后的内容对象
 */
async function fetchZhihuAnswer(url, debug = false) {
  console.log("正在启动浏览器...");
  const browser = await puppeteer.launch({
    headless: debug ? false : true,
    executablePath: os.platform() === 'linux' ? '/usr/bin/chromium-browser' : null,
    args: [
      "--no-sandbox",
      "--disable-setuid-sandbox",
      "--disable-blink-features=AutomationControlled",
      "--disable-dev-shm-usage",
      "--disable-accelerated-2d-canvas",
      "--no-first-run",
      "--no-zygote",
      "--disable-gpu",
      "--lang=zh-CN,zh",
      "--window-size=1920,1080",
      "--start-maximized"
    ]
  });

  try {
    const page = await browser.newPage();

    await page.setUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36");
    
    await page.setExtraHTTPHeaders({
      "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
      "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
      "Accept-Encoding": "gzip, deflate, br",
      "Connection": "keep-alive",
      "Upgrade-Insecure-Requests": "1"
    });

    await page.evaluateOnNewDocument(() => {
      Object.defineProperty(navigator, "webdriver", {
        get: () => false
      });
      
      Object.defineProperty(navigator, "plugins", {
        get: () => [1, 2, 3, 4, 5]
      });
      
      Object.defineProperty(navigator, "languages", {
        get: () => ["zh-CN", "zh", "en"]
      });
      
      window.chrome = {
        runtime: {}
      };
      
      Object.defineProperty(navigator, "permissions", {
        get: () => ({
          query: () => Promise.resolve({ state: "granted" })
        })
      });

      Object.defineProperty(navigator, "platform", {
        get: () => "Win32"
      });

      Object.defineProperty(navigator, "hardwareConcurrency", {
        get: () => 8
      });

      Object.defineProperty(navigator, "deviceMemory", {
        get: () => 8
      });

      const originalQuery = window.navigator.permissions.query;
      window.navigator.permissions.query = (parameters) => (
        parameters.name === "notifications" ?
          Promise.resolve({ state: Notification.permission }) :
          originalQuery(parameters)
      );
    });

    await page.setViewport({ width: 1920, height: 1080 });

    console.log("先访问知乎首页建立会话...");
    await page.goto("https://www.zhihu.com", {
      waitUntil: "networkidle2",
      timeout: 30000,
    });
    await sleep(2000 + Math.random() * 2000);

    await page.evaluate(() => {
      window.scrollTo(0, Math.random() * 300);
    });
    await sleep(1000 + Math.random() * 1000);

    console.log("正在加载目标页面...");
    await page.goto(url, {
      waitUntil: "networkidle2",
      timeout: 1000 * 60 * 3,
      referer: "https://www.zhihu.com"
    });

    await sleep(3000 + Math.random() * 2000);

    const currentUrl = page.url();
    if (currentUrl.includes("unhuman") || currentUrl.includes("account/unhuman")) {
      if (debug) {
        console.log("当前URL:", currentUrl);
        await sleep(10000);
      }
      throw new Error("检测到反爬虫验证页面，请稍后重试或使用登录状态");
    }

    await page.evaluate(() => {
      window.scrollTo(0, Math.random() * 500);
    });
    await sleep(500 + Math.random() * 500);

    console.log("等待内容加载...");
    await page.waitForSelector(".RichContent-inner, .Post-RichText, .RichText", { timeout: 30000 }).catch(() => {
      console.log("警告: 未找到内容元素，尝试继续...");
    });

    await page.evaluate(() => {
      window.scrollTo(0, document.body.scrollHeight / 2);
    });
    await sleep(1000 + Math.random() * 1000);

    // 关闭可能出现的模态框
    console.log("检查并关闭模态框...");
    try {
      const closeButton = await page.$(".Button.Modal-closeButton.Button--plain");
      if (closeButton) {
        await closeButton.click();
        console.log("✓ 已关闭模态框");
        await sleep(500);
      }
    } catch (error) {
      console.log("未找到模态框关闭按钮");
    }

    // 点击展开按钮，显示完整内容
    console.log("检查并展开内容...");
    try {
      const expandButtons = await page.$$(".ContentItem-expandButton");
      if (expandButtons.length > 0) {
        for (const button of expandButtons) {
          await button.click();
          await sleep(500);
        }
        console.log(`✓ 已展开 ${expandButtons.length} 个内容块`);
      }
    } catch (error) {
      console.log("未找到展开按钮");
    }

    await sleep(1000);

    console.log("正在提取内容...");

    // 提取页面数据
    const result = await page.evaluate(() => {
      const data = {
        title: "",
        author: "",
        content: "",
        images: [],
        time: ""
      };

      // 提取标题
      const titleEl = document.querySelector(".QuestionHeader-title");
      if (titleEl) {
        data.title = titleEl.textContent.trim();
      }

      // 提取作者
      let authorEl = document.querySelector(".AuthorInfo-name .UserLink-link");
      if (!authorEl) {
        // 尝试其他可能的选择器
        authorEl = document.querySelector(".AnswerItem-authorInfo .UserLink-link");
      }
      if (!authorEl) {
        // 尝试查找所有 UserLink-link，取第一个包含作者信息的
        const allLinks = document.querySelectorAll(".UserLink-link");
        for (const link of allLinks) {
          if (link.closest(".AuthorInfo") || link.hasAttribute("data-za-detail-view-element_name")) {
            authorEl = link;
            break;
          }
        }
      }
      if (authorEl) {
        data.author = authorEl.textContent.trim();
      }

      // 提取时间信息（包含时间和地点）
      const timeEl = document.querySelector(".ContentItem-time");
      if (timeEl) {
        // 获取完整的时间信息文本（包含地点）
        data.time = timeEl.innerText;
      }

      // 提取回答内容 - 尝试多个可能的选择器
      const contentSelectors = [".RichContent-inner", ".Post-RichText", ".RichText"];
      let contentEl = null;

      for (const selector of contentSelectors) {
        contentEl = document.querySelector(selector);
        if (contentEl) break;
      }

      if (contentEl) {
        // 克隆节点以避免修改原始 DOM
        const clone = contentEl.cloneNode(true);

        // 处理图片 - 提取所有图片 URL
        const figures = clone.querySelectorAll("figure");
        figures.forEach(figure => {
          const img = figure.querySelector("img");
          if (img) {
            let imgUrl = img.getAttribute("data-actualsrc") ||
              img.getAttribute("data-original") ||
              img.getAttribute("src");
            if (imgUrl && (imgUrl.includes("zhimg.com") || imgUrl.includes("pic"))) {
              if (!imgUrl.startsWith("http")) {
                imgUrl = "https:" + imgUrl;
              }
              data.images.push(imgUrl);

              // 查找图片说明文字
              const figcaption = figure.querySelector("figcaption");
              let caption = "";
              if (figcaption) {
                caption = figcaption.textContent.trim();
              }

              // 生成 markdown：图片 + 居中的说明文字 + 两个换行
              let mdText = "\n![](" + imgUrl + ")\n";
              if (caption) {
                mdText += "<center>" + caption + "</center>\n\n";
              }

              const mdImg = document.createTextNode(mdText);
              figure.parentNode.replaceChild(mdImg, figure);
            }
          }
        });

        // 处理不在 figure 中的独立图片
        const images = clone.querySelectorAll("img");
        images.forEach(img => {
          // 跳过已经处理过的（在 figure 中的）
          if (img.closest("figure")) return;

          let imgUrl = img.getAttribute("data-actualsrc") ||
            img.getAttribute("data-original") ||
            img.getAttribute("src");
          if (imgUrl && (imgUrl.includes("zhimg.com") || imgUrl.includes("pic"))) {
            if (!imgUrl.startsWith("http")) {
              imgUrl = "https:" + imgUrl;
            }
            data.images.push(imgUrl);

            // 替换图片标签为 markdown
            const mdImg = document.createTextNode("\n![](" + imgUrl + ")\n");
            img.parentNode.replaceChild(mdImg, img);
          }
        });

        // 处理代码块
        const codeBlocks = clone.querySelectorAll(".highlight");
        codeBlocks.forEach(block => {
          const code = block.querySelector("code");
          if (code) {
            const lang = code.className.replace("language-", "").trim();
            const codeText = code.textContent;
            const mdCode = document.createTextNode("\n```" + lang + "\n" + codeText + "\n```\n");
            block.parentNode.replaceChild(mdCode, block);
          }
        });

        // 处理知乎智能词条链接（只保留文字，去掉链接）
        const entityWords = clone.querySelectorAll(".RichContent-EntityWord");
        entityWords.forEach(link => {
          // 移除SVG图标
          const svg = link.querySelector("svg");
          if (svg) {
            svg.remove();
          }
          // 只保留文本内容
          const text = link.textContent.trim();
          const textNode = document.createTextNode(text);
          link.parentNode.replaceChild(textNode, link);
        });

        // 处理链接卡片
        const linkCards = clone.querySelectorAll(".RichText-LinkCardContainer");
        linkCards.forEach(card => {
          const link = card.querySelector("a");
          if (link) {
            const text = link.getAttribute("data-text") || link.textContent;
            const href = link.href;
            const mdLink = document.createTextNode("\n[" + text + "](" + href + ")\n");
            card.parentNode.replaceChild(mdLink, card);
          }
        });

        // 处理视频
        const videos = clone.querySelectorAll("video");
        videos.forEach(video => {
          const src = video.getAttribute("src");
          if (src) {
            const mdVideo = document.createTextNode("\n[视频](" + src + ")\n");
            video.parentNode.replaceChild(mdVideo, video);
          }
        });

        // 处理标题
        clone.querySelectorAll("h2").forEach(h => {
          const text = h.textContent;
          const md = document.createTextNode("\n## " + text + "\n");
          h.parentNode.replaceChild(md, h);
        });

        clone.querySelectorAll("h3").forEach(h => {
          const text = h.textContent;
          const md = document.createTextNode("\n### " + text + "\n");
          h.parentNode.replaceChild(md, h);
        });

        // 处理加粗
        clone.querySelectorAll("b, strong").forEach(el => {
          const text = el.textContent;
          const md = document.createTextNode("**" + text + "**");
          el.parentNode.replaceChild(md, el);
        });

        // 处理斜体
        clone.querySelectorAll("i, em").forEach(el => {
          const text = el.textContent;
          const md = document.createTextNode("*" + text + "*");
          el.parentNode.replaceChild(md, el);
        });

        // 处理行内代码
        clone.querySelectorAll("code").forEach(el => {
          const text = el.textContent;
          const md = document.createTextNode("`" + text + "`");
          el.parentNode.replaceChild(md, el);
        });

        // 处理链接
        clone.querySelectorAll("a").forEach(el => {
          const text = el.textContent;
          const href = el.href;
          const md = document.createTextNode("[" + text + "](" + href + ")");
          el.parentNode.replaceChild(md, el);
        });

        // 处理列表
        clone.querySelectorAll("li").forEach(el => {
          const text = el.textContent;
          const md = document.createTextNode("- " + text + "\n");
          el.parentNode.replaceChild(md, el);
        });

        // 处理引用
        clone.querySelectorAll("blockquote").forEach(el => {
          const text = el.textContent.trim().split("\n").map(line => "> " + line).join("\n");
          const md = document.createTextNode("\n" + text + "\n");
          el.parentNode.replaceChild(md, el);
        });

        // 处理段落和换行
        clone.querySelectorAll("br").forEach(el => {
          const md = document.createTextNode("\n");
          el.parentNode.replaceChild(md, el);
        });

        clone.querySelectorAll("p").forEach(el => {
          const text = el.textContent;
          if (text.trim()) {
            const md = document.createTextNode("\n" + text + "\n");
            el.parentNode.replaceChild(md, el);
          }
        });

        // 获取最终的文本内容
        data.content = clone.textContent
          .replace(/\n{3,}/g, "\n\n")
          .trim();
      }

      return data;
    });

    return { result, browser };
  } catch (error) {
    await browser.close();
    throw error;
  }
}


/**
 * 爬取知乎回答并保存为 Markdown 文件
 * @param {string} url - 知乎回答链接
 * @param {boolean} debugMode - 是否开启调试模式
 */
export async function crawlAndSaveZhihuAnswer(url, debugMode = false) {
  let browser = null;

  try {
    console.log(`开始爬取: ${url}`);
    console.log("使用 puppeteer-extra + stealth 插件绕过反爬虫检测");
    if (debugMode) console.log("🐛 调试模式已开启");

    // 🕷 获取并解析页面内容
    const data = await fetchZhihuAnswer(url, debugMode);
    const result = data.result;
    browser = data.browser;

    // 📂 创建输出目录
    const outputDir = path.join(__dirname, "output");
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }

    // 📝 生成文件名
    const filenameParts = [];
    if (result.title) filenameParts.push(sanitizeFilename(result.title));
    if (result.author) filenameParts.push(sanitizeFilename(result.author));

    const filename =
      filenameParts.length > 0
        ? `${filenameParts.join(" - ")}.md`
        : `zhihu_answer_${Date.now()}.md`;
    const filepath = path.join(outputDir, filename);

    // 🧱 构建 Markdown 内容
    let markdown = result.content + "\n\n";

    if (result.time) markdown += `${result.time}\n\n`;

    // 添加原文链接
    let linkText = "";
    if (result.title) linkText += result.title;
    if (result.author) linkText += ` - ${result.author}的回答`;
    if (linkText) linkText += " - 知乎";
    markdown += `[${linkText || "原文链接"}](${url})\n`;

    // 💾 写入文件
    fs.writeFileSync(filepath, markdown, "utf-8");
    console.log(`\n✓ Markdown 文件已保存: ${filepath}`);

    // ✅ 输出统计信息
    console.log("\n爬取完成！");
    console.log(`标题: ${result.title || "(未找到)"}`);
    console.log(`作者: ${result.author || "(未找到)"}`);
    console.log(`内容长度: ${result.content.length} 字符`);
    console.log(`图片数量: ${result.images.length} 张（使用知乎原始链接）`);

    // 🧩 调试模式保持浏览器
    if (debugMode) {
      console.log("\n🐛 调试模式：浏览器保持打开状态，按 Ctrl+C 退出");
      await new Promise(() => { });
    } else if (browser) {
      await browser.close();
    }

  } catch (error) {
    console.error("❌ 爬取失败:", error.message);
    console.error(error.stack);
    if (browser) await browser.close();
    throw error; // 交给上层处理（比如 withFileLock）
  }
}

async function main() {
  const args = process.argv.slice(2);
  const debugMode = args.includes("--debug");
  const url = args.find(arg => !arg.startsWith("--"));

  if (!url) {
    console.error("请提供要爬取的 URL, 例如：");
    console.error("node crawler.js https://www.zhihu.com/question/xxxx/answer/xxxx");
    process.exit(1);
  }

  const lockPath = path.join(__dirname, "crawler.lock");

  await withFileLock(lockPath, async () => {
    await crawlAndSaveZhihuAnswer(url, debugMode);
  });
}

main();

