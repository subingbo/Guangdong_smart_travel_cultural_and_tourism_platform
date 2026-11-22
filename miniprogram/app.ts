App<IAppOption>({
  globalData: {},
  onLaunch() {
    // 记录启动日志（保留原有能力）
    const logs = (wx.getStorageSync('logs') || []) as number[]
    logs.unshift(Date.now())
    wx.setStorageSync('logs', logs)
  },
})
