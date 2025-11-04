import fs from "fs";
import path from "path";

/**
 * 阻塞等待文件锁（带自动清理）
 * @param {string} lockPath 锁文件路径
 * @param {Function} taskFn 要执行的异步任务
 */
export async function withFileLock(lockPath, taskFn) {
  const dir = path.dirname(lockPath);
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
  setupCleanup(lockPath);

  const waitInterval = 1000; // 每隔 1 秒检测一次
  let hasLock = false;

  try {
    // 等待锁释放
    while (fs.existsSync(lockPath)) {
      console.log(`检测到其他进程正在执行，等待锁释放...`);
      await new Promise((resolve) => setTimeout(resolve, waitInterval));
    }

    // 创建锁文件
    fs.writeFileSync(lockPath, String(process.pid));
    hasLock = true;
    console.log("已获取执行锁，开始任务...");

    await taskFn();
  } catch (err) {
    console.error("执行任务失败:", err);
  } finally {
    releaseLock(lockPath);
  }
}

const releaseLock = (lockPath) => {
  if (fs.existsSync(lockPath)) {
    try {
      fs.unlinkSync(lockPath);
      console.log("锁已释放。");
    } catch (err) {
      console.warn("删除锁文件失败:", err.message);
    }
  }
};

// 注册清理逻辑
const setupCleanup = (lockPath) => {
  const cleanAndExit = (code) => {
    releaseLock(lockPath);
    process.exit(code ?? 0);
  };

  process.on("exit", () => releaseLock(lockPath));
  process.on("SIGINT", () => cleanAndExit(0)); // Ctrl+C
  process.on("SIGTERM", () => cleanAndExit(0));
  process.on("uncaughtException", (err) => {
    console.error("未捕获异常:", err);
    cleanAndExit(1);
  });
};

