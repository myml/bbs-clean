# 深度论坛清理广告贴

针对深度社区论坛的恶意广告、灌水贴，会自动禁言用户并隐藏用户的发帖内容

触发条件：

1. 贴子内容包含过多链接（链接数大于100个）
2. 连续重复发帖（同一个用户有两个以上的贴子在首页）
3. 贴子标题包含黑名单关键字
4. 根据历史贴子数据训练的机器学习模型

仅对论坛等级低于三级的用户生效果。

添加关键词见 [keywords.go](./keywords.go) 文件

