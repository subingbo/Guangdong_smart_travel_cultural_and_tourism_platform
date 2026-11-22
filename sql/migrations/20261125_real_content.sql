-- Realistic demo content for the website (run once on an existing DB).
-- Fresh installs get the same content from sql/schema.sql + sql/seed.sql.
SET NAMES utf8mb4;

ALTER TABLE products
  ADD COLUMN sold_count INT NOT NULL DEFAULT 0 AFTER stock,
  ADD COLUMN rating DECIMAL(2,1) NOT NULL DEFAULT 5.0 AFTER sold_count,
  ADD COLUMN review_count INT NOT NULL DEFAULT 0 AFTER rating;

CREATE TABLE IF NOT EXISTS product_reviews (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  product_id BIGINT UNSIGNED NOT NULL,
  user_id    BIGINT UNSIGNED NOT NULL,
  rating     TINYINT         NOT NULL DEFAULT 5,
  content    VARCHAR(1000)   NOT NULL,
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_product_reviews_product (product_id, id),
  CONSTRAINT fk_product_reviews_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
  CONSTRAINT fk_product_reviews_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS activities (
  id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  title_cn       VARCHAR(160)    NOT NULL,
  title_en       VARCHAR(200)    NOT NULL,
  description_cn TEXT            NOT NULL,
  description_en TEXT            NOT NULL,
  location       VARCHAR(160)    NOT NULL,
  start_at       DATETIME        NOT NULL,
  end_at         DATETIME        NOT NULL,
  image_url      VARCHAR(500)    NOT NULL DEFAULT '',
  organizer      VARCHAR(160)    NOT NULL DEFAULT '',
  tag            VARCHAR(50)     NOT NULL DEFAULT '',
  status         VARCHAR(20)     NOT NULL DEFAULT 'upcoming',
  created_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_activities_start (start_at),
  KEY idx_activities_status (status, start_at)
) ENGINE = InnoDB;

-- Scenic images: replace generic Unsplash landscapes with real Guangdong photos.
UPDATE scenic_images SET image_url = '/seed/yuexiu.jpg',
  title_cn = '越秀公园晨景', title_en = 'Morning at Yuexiu Park',
  description_cn = '清晨的越秀公园，五羊石像与镇海楼在树影里慢慢醒来。',
  description_en = 'A quiet morning at Yuexiu Park, with the Five Rams Statue and Zhenhai Tower.'
  WHERE sort_order = 1;
UPDATE scenic_images SET image_url = '/seed/pearl-river.jpg',
  title_cn = '珠江夜景', title_en = 'Pearl River at Night',
  description_cn = '广州塔亮灯之后，珠江两岸像一条流动的光带。',
  description_en = 'After Canton Tower lights up, the Pearl River becomes a ribbon of light.'
  WHERE sort_order = 2;
UPDATE scenic_images SET image_url = '/seed/chen-clan.jpg',
  title_cn = '陈家祠古韵', title_en = 'Chen Clan Academy',
  description_cn = '岭南建筑的代表之作，木雕、砖雕与灰塑一次看全。',
  description_en = 'A masterpiece of Lingnan architecture with wood, brick and clay carvings.'
  WHERE sort_order = 3;
UPDATE scenic_images SET image_url = '/seed/shamian.jpg',
  title_cn = '沙面岛午后', title_en = 'Shamian Island Afternoon',
  description_cn = '欧式建筑与骑楼交织，适合慢慢走、拍照和喝一杯咖啡。',
  description_en = 'European-style buildings and arcades, perfect for a slow walk and coffee.'
  WHERE sort_order = 4;
UPDATE scenic_images SET image_url = '/seed/dim-sum.jpg',
  title_cn = '一盅两件', title_en = 'Cantonese Morning Tea',
  description_cn = '虾饺、烧卖、叉烧包，一壶茶就是一个广州早晨。',
  description_en = 'Shrimp dumplings, siu mai and BBQ pork buns — a Guangzhou morning in a teapot.'
  WHERE sort_order = 5;

-- Realistic community accounts (password for all: demo123).
UPDATE users SET nickname = '林晓岚', avatar_url = 'https://i.pravatar.cc/150?img=47' WHERE username = 'demo';
UPDATE users SET nickname = '陈子昂', avatar_url = 'https://i.pravatar.cc/150?img=12' WHERE username = 'alice';

INSERT INTO users (username, password_hash, nickname, avatar_url, language, role) VALUES
('maya',  '$2a$10$WxZbvgVv.iDF7ursmeq4n.W.ZF2zFpCo84Ah/Q6QAZrbksnmyuODu', 'Maya Chen', 'https://i.pravatar.cc/150?img=44', 'en', 'user'),
('kenji', '$2a$10$WxZbvgVv.iDF7ursmeq4n.W.ZF2zFpCo84Ah/Q6QAZrbksnmyuODu', 'Kenji Sato', 'https://i.pravatar.cc/150?img=33', 'en', 'user'),
('lena',  '$2a$10$WxZbvgVv.iDF7ursmeq4n.W.ZF2zFpCo84Ah/Q6QAZrbksnmyuODu', 'Lena Müller', 'https://i.pravatar.cc/150?img=45', 'en', 'user'),
('jiaqi', '$2a$10$WxZbvgVv.iDF7ursmeq4n.W.ZF2zFpCo84Ah/Q6QAZrbksnmyuODu', '黄嘉怡', 'https://i.pravatar.cc/150?img=32', 'zh', 'user'),
('qiming','$2a$10$WxZbvgVv.iDF7ursmeq4n.W.ZF2zFpCo84Ah/Q6QAZrbksnmyuODu', '周启明', 'https://i.pravatar.cc/150?img=15', 'zh', 'user'),
('sofia', '$2a$10$WxZbvgVv.iDF7ursmeq4n.W.ZF2zFpCo84Ah/Q6QAZrbksnmyuODu', 'Sofia Rossi', 'https://i.pravatar.cc/150?img=49', 'en', 'user')
ON DUPLICATE KEY UPDATE nickname = VALUES(nickname), avatar_url = VALUES(avatar_url), role = 'user';

SET @demo := (SELECT id FROM users WHERE username = 'demo');
SET @alice := (SELECT id FROM users WHERE username = 'alice');
SET @maya := (SELECT id FROM users WHERE username = 'maya');
SET @kenji := (SELECT id FROM users WHERE username = 'kenji');
SET @lena := (SELECT id FROM users WHERE username = 'lena');
SET @jiaqi := (SELECT id FROM users WHERE username = 'jiaqi');
SET @qiming := (SELECT id FROM users WHERE username = 'qiming');
SET @sofia := (SELECT id FROM users WHERE username = 'sofia');

-- Reset editorial/social demo data and rebuild a believable community feed.
DELETE FROM notifications;
DELETE FROM follows;
DELETE FROM blog_likes;
DELETE FROM comments;
DELETE FROM travel_blogs;

INSERT INTO travel_blogs (id, user_id, title, content, images, likes, created_at) VALUES
(1, @demo, '在陈家祠，看了一下午木雕',
 '以前路过陈家祠很多次，这次终于慢慢走完。屋脊上的灰塑、梁架上的木雕、窗棂的满洲窗，每一处都值得停下来。最喜欢聚贤堂前的那排木雕，阳光斜下来时，人物和花鸟像活了一样。建议下午三点后去，人少一些，光线也柔。',
 '["/seed/chen-clan.jpg"]', 128, '2026-11-08 15:20:00'),
(2, @maya, 'A Slow Morning Tea in Liwan',
 'Woke up early for yum cha at an old teahouse in Liwan. Shrimp dumplings, siu mai, chicken feet and a pot of pu''er. The auntie at the next table taught me to tap two fingers to say thank you. A slow morning, exactly how Guangzhou should start.',
 '["/seed/dim-sum.jpg"]', 96, '2026-11-10 09:45:00'),
(3, @alice, '珠江夜游：广州的夜从水面开始',
 '从海心沙上船，江风一吹，广州塔的灯刚好亮起来。两岸的楼宇像一排排流动的屏幕，猎德大桥从头顶掠过。60 分钟不长，但足够把这座城市的夜看一遍。',
 '["/seed/pearl-river.jpg"]', 214, '2026-11-12 21:10:00'),
(4, @lena, 'Shamian Island: An Afternoon in Old Canton',
 'Shamian feels like a film set: colonial arcades, banyan trees, church bells and quiet side streets. We walked without a map, stopped for coffee, and watched the light change on the Pearl River. A perfect slow afternoon.',
 '["/seed/shamian.jpg"]', 76, '2026-11-15 16:30:00'),
(5, @jiaqi, '永庆坊的骑楼、咖啡和一场粤剧',
 '永庆坊的骑楼修旧如旧，转角是咖啡店，巷子里却还能听到粤剧。周六下午遇到一场街坊粤剧，阿叔拉高胡，阿姨唱《帝女花》，游客和街坊围了一圈。广州的文化不在展柜里，就在这些日常里。',
 '["/seed/cantonese-opera.jpg"]', 143, '2026-11-17 17:05:00'),
(6, @demo, '周末夜游珠江：从海心沙走到广州塔',
 '周末不想挤地铁，就从海心沙一路走到广州塔。江边有人跑步、有人遛狗、有人架着三脚架拍夜景。走到广州塔下抬头看，才发现它比照片里更高。',
 '["/seed/pearl-river.jpg"]', 58, '2026-11-19 20:40:00'),
(7, @kenji, 'Lion Dance, Tea and the Taste of Foshan',
 'Took the metro to Foshan for the lion dance at the ancestral temple. The drums started, the lion blinked, and the whole courtyard went quiet. Afterwards, tea and a bowl of wonton noodles. A day trip worth every minute.',
 '["/seed/lion-dance.jpg"]', 88, '2026-11-20 14:15:00'),
(8, @qiming, '给第一次来广州的朋友：一条不走回头路的早茶路线',
 '第一次来广州，早茶不要一次点太多。我的路线是：早上先到泮溪酒家，点虾饺、烧卖、艇仔粥；十点去陈家祠消食；中午在长寿路吃一碗云吞面；下午去沙面岛喝咖啡。一天刚刚好。',
 '["/seed/morning-tea.jpg"]', 64, '2026-11-21 11:30:00');

INSERT INTO comments (blog_id, user_id, content, created_at) VALUES
(1, @maya, '木雕的细节拍得真好，聚贤堂那排我也站了很久。', '2026-11-08 16:02:00'),
(1, @jiaqi, '下次可以顺路去旁边的广东民间工艺博物馆，很多广绣藏品。', '2026-11-08 18:10:00'),
(1, @demo, '谢谢推荐，下次一定去。', '2026-11-08 18:22:00'),
(2, @demo, 'Tap two fingers is such a lovely detail. Adding this teahouse to my list.', '2026-11-10 10:20:00'),
(2, @lena, 'The chicken feet look intimidating but taste amazing. 😂', '2026-11-10 11:05:00'),
(2, @maya, 'Haha, the auntie was very patient with me.', '2026-11-10 11:12:00'),
(3, @kenji, 'The Canton Tower lights are incredible. Which pier did you leave from?', '2026-11-12 21:35:00'),
(3, @alice, '天字码头，晚上七点半那班，建议提前 20 分钟到。', '2026-11-12 21:48:00'),
(3, @sofia, 'Saving this for my Guangzhou trip next month!', '2026-11-13 08:30:00'),
(4, @demo, '沙面的猫也很多，带点猫粮会被围住。', '2026-11-15 17:10:00'),
(4, @lena, 'I saw three cats sleeping under a banyan tree. 😻', '2026-11-15 17:25:00'),
(5, @qiming, '永庆坊现在周末人不少，建议早上十点前到。', '2026-11-17 18:02:00'),
(5, @jiaqi, '是的，早上去拍照舒服很多。', '2026-11-17 18:15:00'),
(6, @maya, 'The walk from Haixinsha is so worth it at night.', '2026-11-19 21:05:00'),
(6, @demo, '下次试试从猎德桥走到琶醍。', '2026-11-19 21:20:00'),
(7, @jiaqi, '佛山的醒狮真的很有气势，鼓点一响鸡皮疙瘩都起来了。', '2026-11-20 15:00:00'),
(7, @kenji, 'Exactly! The drums were amazing.', '2026-11-20 15:18:00'),
(8, @lena, 'This is a perfect one-day plan. Do you need to book the teahouse?', '2026-11-21 12:10:00'),
(8, @qiming, '周末最好提前订位，平日直接去也行。', '2026-11-21 12:25:00');

INSERT IGNORE INTO blog_likes (user_id, blog_id) VALUES
(@demo, 2), (@demo, 3), (@demo, 5), (@demo, 7),
(@maya, 1), (@maya, 6), (@alice, 2), (@alice, 7),
(@kenji, 1), (@kenji, 3), (@lena, 2), (@jiaqi, 3), (@qiming, 6);

INSERT IGNORE INTO follows (follower_id, followee_id) VALUES
(@demo, @maya), (@demo, @lena), (@demo, @jiaqi), (@demo, @kenji),
(@maya, @demo), (@maya, @jiaqi),
(@alice, @demo), (@alice, @kenji),
(@lena, @demo), (@lena, @maya),
(@jiaqi, @demo), (@jiaqi, @maya),
(@kenji, @demo), (@kenji, @alice),
(@qiming, @demo);

INSERT INTO notifications (user_id, actor_id, type, blog_id, comment_id, is_read, created_at) VALUES
(@demo, @maya, 'like', 1, NULL, 0, '2026-11-08 16:00:00'),
(@demo, @jiaqi, 'comment', 1, (SELECT id FROM comments WHERE blog_id = 1 AND user_id = @jiaqi LIMIT 1), 0, '2026-11-08 18:10:00'),
(@demo, @lena, 'follow', NULL, NULL, 0, '2026-11-15 17:30:00'),
(@demo, @kenji, 'like', 3, NULL, 1, '2026-11-13 09:10:00'),
(@demo, @maya, 'comment', 6, (SELECT id FROM comments WHERE blog_id = 6 AND user_id = @maya LIMIT 1), 1, '2026-11-19 21:05:00');

-- Product catalogue: real photos, sales/rating metadata.
UPDATE products SET name_cn='陈家祠·岭南建筑导览票', name_en='Chen Clan Academy Heritage Ticket',
  description_cn='含陈家祠门票与岭南建筑导览手册，木雕、砖雕、灰塑一次看全。',
  description_en='Admission plus a Lingnan architecture guide; wood, brick and clay carvings in one visit.',
  price_cents=1000, category='ticket', image_url='/seed/chen-clan.jpg', stock=500, sold_count=1286, rating=4.8, review_count=326
  WHERE name_cn='陈家祠门票';
UPDATE products SET name_cn='珠江夜游·水晶号船票', name_en='Pearl River Night Cruise · Crystal',
  description_cn='天字码头登船，沿江看广州塔与两岸灯光，约 60 分钟。',
  description_en='Board at Tianzi Pier for a 60-minute illuminated cruise past Canton Tower.',
  price_cents=12800, category='ticket', image_url='/seed/pearl-river.jpg', stock=200, sold_count=842, rating=4.9, review_count=517
  WHERE name_cn='珠江夜游船票';
UPDATE products SET name_cn='西关早茶双人套餐', name_en='Xiguan Dim Sum for Two',
  description_cn='虾饺、烧卖、叉烧包、豉汁凤爪与一壶普洱，两个人刚刚好。',
  description_en='Shrimp dumplings, siu mai, BBQ pork buns, chicken feet and a pot of pu''er for two.',
  price_cents=8800, category='snack', image_url='/seed/dim-sum.jpg', stock=120, sold_count=2134, rating=4.7, review_count=689
  WHERE name_cn='广式早茶双人套餐';
UPDATE products SET name_cn='岭南醒狮文创盲盒', name_en='Lingnan Lion Dance Blind Box',
  description_cn='随机一款醒狮、骑楼或粤剧主题文创，附双语文化卡片。',
  description_en='A surprise Lingnan keepsake with a bilingual culture card.',
  price_cents=5900, category='souvenir', image_url='/seed/lion-dance.jpg', stock=300, sold_count=956, rating=4.6, review_count=208
  WHERE name_cn='岭南文创盲盒';
UPDATE products SET name_cn='手打香水柠檬茶', name_en='Hand-smashed Lemon Tea',
  description_cn='香水柠檬现捶，茶底清爽，广式茶饮的夏日代表。',
  description_en='Freshly smashed fragrant lemon over chilled tea, a Cantonese summer classic.',
  price_cents=1800, category='snack', image_url='/seed/lemon-tea.jpg', stock=999, sold_count=3120, rating=4.5, review_count=412
  WHERE name_cn='手打柠檬茶';
UPDATE products SET name_cn='沙面岛·欧陆建筑旅拍', name_en='Shamian Island Photo Walk',
  description_cn='本地摄影师带你走沙面欧陆建筑群，含 9 张精修照片。',
  description_en='A guided photo walk through Shamian''s European-style streets, with 9 edited photos.',
  price_cents=29900, category='experience', image_url='/seed/shamian.jpg', stock=40, sold_count=126, rating=4.9, review_count=48
  WHERE name_cn='沙面岛旅拍体验';
UPDATE products SET name_cn='广彩瓷·花果纹茶杯', name_en='Canton Porcelain Floral Cup',
  description_cn='手工广彩，花果纹样，配礼盒包装，适合作为岭南伴手礼。',
  description_en='Hand-painted Canton porcelain with floral motifs, gift-boxed.',
  price_cents=16800, category='souvenir', image_url='/seed/porcelain.jpg', stock=80, sold_count=318, rating=4.8, review_count=76
  WHERE name_cn='广彩瓷杯';
UPDATE products SET name_cn='周末文旅消费券', name_en='Weekend Culture Voucher',
  description_cn='满 100 减 20，适用于平台内门票、体验与文创商品。',
  description_en='Spend 100, save 20 on tickets, experiences and crafts across the platform.',
  price_cents=0, category='coupon', image_url='/seed/cantonese-opera.jpg', stock=1000, sold_count=5200, rating=4.9, review_count=860
  WHERE name_cn='周末文旅消费券';

-- Product reviews (selected quotes; aggregate review_count stays larger).
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @maya, 5, '陈家祠的木雕和灰塑比想象中更精彩，导览手册也很实用。', '2026-11-09 10:20:00' FROM products WHERE name_en='Chen Clan Academy Heritage Ticket';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @jiaqi, 5, '带外地朋友来的，讲解很仔细，下午的光线特别适合拍照。', '2026-11-12 15:40:00' FROM products WHERE name_en='Chen Clan Academy Heritage Ticket';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @kenji, 4, 'A must-visit for anyone interested in Lingnan architecture.', '2026-11-16 09:10:00' FROM products WHERE name_en='Chen Clan Academy Heritage Ticket';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @lena, 5, '夜景很值，建议选七点半那班，江风很舒服。', '2026-11-13 21:30:00' FROM products WHERE name_en='Pearl River Night Cruise · Crystal';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @sofia, 5, 'The skyline from the water is stunning. Book the earlier slot!', '2026-11-14 20:05:00' FROM products WHERE name_en='Pearl River Night Cruise · Crystal';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @demo, 4, '船很稳，甲板视角好，记得带一件薄外套。', '2026-11-18 21:15:00' FROM products WHERE name_en='Pearl River Night Cruise · Crystal';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @maya, 5, '虾饺皮薄馅足，凤爪入味，两个人吃刚好。', '2026-11-11 10:05:00' FROM products WHERE name_en='Xiguan Dim Sum for Two';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @qiming, 5, '周末带家人去的，点心现做，服务也很热情。', '2026-11-16 11:20:00' FROM products WHERE name_en='Xiguan Dim Sum for Two';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @lena, 4, 'A great introduction to dim sum. Come hungry!', '2026-11-20 12:00:00' FROM products WHERE name_en='Xiguan Dim Sum for Two';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @jiaqi, 5, '抽到了醒狮款，做工比预期好，卡片也很有文化感。', '2026-11-10 18:30:00' FROM products WHERE name_en='Lingnan Lion Dance Blind Box';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @sofia, 4, 'Cute souvenir, perfect small gift for friends back home.', '2026-11-17 16:45:00' FROM products WHERE name_en='Lingnan Lion Dance Blind Box';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @demo, 5, '不会太甜，柠檬味很足，夏天喝很爽。', '2026-11-19 14:20:00' FROM products WHERE name_en='Hand-smashed Lemon Tea';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @kenji, 4, 'Refreshing and not too sweet. Perfect after a long walk.', '2026-11-21 15:10:00' FROM products WHERE name_en='Hand-smashed Lemon Tea';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @lena, 5, '摄影师很会找角度，沙面的光线太美了。', '2026-11-15 18:10:00' FROM products WHERE name_en='Shamian Island Photo Walk';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @maya, 5, 'Loved the route and the edited photos. Highly recommend.', '2026-11-18 19:00:00' FROM products WHERE name_en='Shamian Island Photo Walk';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @qiming, 5, '杯子很精致，包装也稳，送人合适。', '2026-11-16 10:40:00' FROM products WHERE name_en='Canton Porcelain Floral Cup';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @sofia, 4, 'Beautiful craftsmanship; the gift box is a nice touch.', '2026-11-20 17:25:00' FROM products WHERE name_en='Canton Porcelain Floral Cup';
INSERT INTO product_reviews (product_id, user_id, rating, content, created_at)
SELECT id, @demo, 5, '周末去陈家祠直接抵扣，很方便。', '2026-11-21 09:30:00' FROM products WHERE name_en='Weekend Culture Voucher';

-- Upcoming cultural activities.
DELETE FROM activities;
INSERT INTO activities (title_cn, title_en, description_cn, description_en, location, start_at, end_at, image_url, organizer, tag, status) VALUES
('广州国际美食节·岭南风味周', 'Guangzhou International Food Festival', '连续十天，把广府、潮汕、客家风味一次吃遍。现场有老字号摊位、点心手作和厨艺分享。', 'Ten days of Cantonese, Chaoshan and Hakka flavors, with heritage stalls, dim sum workshops and chef talks.', '天河体育中心', '2026-11-28 10:00:00', '2026-12-06 22:00:00', '/seed/dim-sum.jpg', '广州市文化广电旅游局', '美食', 'upcoming'),
('陈家祠非遗手作周', 'Chen Clan Heritage Craft Week', '广绣、广彩、灰塑、木雕传承人现场演示，周末还有亲子手作体验。', 'Live demonstrations by Canton embroidery, porcelain, clay sculpture and wood carving artisans, with weekend family workshops.', '陈家祠', '2026-12-05 09:00:00', '2026-12-13 17:30:00', '/seed/chen-clan.jpg', '广东民间工艺博物馆', '非遗', 'upcoming'),
('珠江夜游·城市灯光季', 'Pearl River Lights Season', '天字码头出发，串联广州塔、猎德大桥与琶醍，船上还有粤语讲解。', 'Depart from Tianzi Pier for Canton Tower, Liede Bridge and Party Pier, with Cantonese commentary on board.', '天字码头', '2026-11-25 19:00:00', '2026-12-31 22:30:00', '/seed/pearl-river.jpg', '珠江夜游', '夜游', 'upcoming'),
('沙面岛周末摄影漫步', 'Shamian Weekend Photo Walk', '本地摄影师带队，走沙面欧陆建筑群与江边，适合手机和相机用户。', 'A local photographer leads a walk through Shamian''s colonial streets and riverfront. Phones and cameras welcome.', '沙面岛', '2026-12-12 15:00:00', '2026-12-12 18:00:00', '/seed/shamian.jpg', '岭南影像社', '摄影', 'upcoming'),
('粤剧文化体验日', 'Cantonese Opera Experience Day', '折子戏、脸谱体验与后台探班，适合第一次接触粤剧的观众。', 'Highlights from Cantonese opera, mask painting and a backstage tour for first-time audiences.', '广州粤剧院', '2026-12-20 14:00:00', '2026-12-20 17:00:00', '/seed/cantonese-opera.jpg', '广州粤剧院', '粤剧', 'upcoming');
