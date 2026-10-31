-- Travel Guangdong - seed data (tourist flow)
USE travel_gd;

-- Demo user: username "demo", password "demo123" (bcrypt hash)
INSERT INTO users (username, password_hash, nickname, avatar_url, language) VALUES
('demo', '$2a$10$WxZbvgVv.iDF7ursmeq4n.W.ZF2zFpCo84Ah/Q6QAZrbksnmyuODu', 'Traveler A', 'https://i.pravatar.cc/150?img=1', 'en'),
('alice', '$2a$10$WxZbvgVv.iDF7ursmeq4n.W.ZF2zFpCo84Ah/Q6QAZrbksnmyuODu', 'Traveler B', 'https://i.pravatar.cc/150?img=2', 'en');

INSERT INTO scenic_images (image_url, title_cn, title_en, location, tag, description_cn, description_en, sort_order) VALUES
('https://images.unsplash.com/photo-1500530855697-b586d89ba3ee', '越秀山晨景', 'Morning at Yuexiu Hill', 'Guangzhou', 'nature', '清晨的越秀山空气清新，适合慢跑与散步。', 'Fresh morning air at Yuexiu Hill, perfect for jogging and strolling.', 1),
('https://images.unsplash.com/photo-1500534314209-a25ddb2bd429', '珠江夜景', 'Pearl River Night View', 'Guangzhou', 'city', '珠江两岸灯光璀璨，是广州夜游必看。', 'Brilliant lights along the Pearl River, a must-see night view.', 2),
('https://images.unsplash.com/photo-1500534623283-312aade485b7', '陈家祠古韵', 'Chen Clan Ancestral Hall', 'Guangzhou', 'culture', '岭南建筑的代表之作，雕刻精美。', 'A masterpiece of Lingnan architecture with exquisite carvings.', 3),
('https://images.unsplash.com/photo-1500530855697-b586d89ba3ee', '沙面岛午后', 'Shamian Island Afternoon', 'Guangzhou', 'heritage', '欧式建筑与骑楼交织，适合拍照打卡。', 'European-style buildings and arcades, great for photos.', 4),
('https://images.unsplash.com/photo-1500534314209-a25ddb2bd429', '广州塔远眺', 'Canton Tower View', 'Guangzhou', 'landmark', '城市地标，视野开阔，夜景更迷人。', 'City landmark with expansive views and stunning night lights.', 5);

INSERT INTO places (name_cn, name_en, lat, lng, tags, indoor, duration, description_cn, description_en) VALUES
('越秀公园', 'Yuexiu Park', 23.1387000, 113.2644000, '["park","history"]', 0, 120, '广州最大的综合性公园之一。', 'One of the largest comprehensive parks in Guangzhou.'),
('陈家祠', 'Chen Clan Ancestral Hall', 23.1286000, 113.2514000, '["culture","architecture"]', 1, 90, '岭南建筑代表，民间工艺展示。', 'Representative Lingnan architecture with folk arts.'),
('沙面岛', 'Shamian Island', 23.1146000, 113.2488000, '["heritage","photo"]', 0, 60, '欧陆风情街区，适合漫步。', 'European-style district ideal for walking.'),
('广州塔', 'Canton Tower', 23.1065000, 113.3247000, '["landmark","view"]', 0, 120, '广州城市地标，登塔观景。', 'Guangzhou landmark with observation deck.'),
('上下九步行街', 'Shangxiajiu Pedestrian Street', 23.1259000, 113.2443000, '["shopping","food"]', 0, 90, '岭南风情商业街，美食云集。', 'Lingnan-style commercial street with local food.');

INSERT INTO knowledge (title_cn, title_en, content_cn, content_en, tags) VALUES
('醒狮文化', 'Lion Dance', '醒狮是岭南传统民俗，寓意辟邪纳福。', 'Lion dance is a Lingnan folk tradition symbolizing good luck.', '["culture","festival"]'),
('广绣', 'Guangzhou Embroidery', '广绣色彩鲜艳、层次丰富，是中国四大名绣之一。', 'Guangzhou embroidery is vivid and layered, one of China''s four famous embroideries.', '["craft","heritage"]'),
('早茶文化', 'Dim Sum Culture', '饮早茶是广州人的生活方式，讲究一盅两件。', 'Dim sum is a lifestyle in Guangzhou, known for tea and small dishes.', '["food","culture"]'),
('岭南建筑', 'Lingnan Architecture', '岭南建筑注重通风采光，常见骑楼和满洲窗。', 'Lingnan architecture focuses on ventilation and light, featuring arcades and Manchurian windows.', '["architecture","history"]'),
('粤剧', 'Cantonese Opera', '粤剧唱腔独特，融合戏曲、音乐、武术。', 'Cantonese opera blends drama, music, and martial arts with distinctive singing.', '["performance","culture"]');

INSERT INTO travel_blogs (user_id, title, content, images, likes) VALUES
(1, 'A Day in Yuexiu Park', 'The park is huge and beautiful. I saw the Five Rams Statue.', '["https://images.unsplash.com/photo-1500530855697-b586d89ba3ee"]', 12),
(2, 'Delicious Dim Sum', 'Tried chicken feet and shrimp dumplings. Absolutely loved it!', '["https://images.unsplash.com/photo-1500534314209-a25ddb2bd429"]', 25),
(2, 'Night Cruise on Pearl River', 'The lights are amazing. Highly recommend taking the boat tour.', '["https://images.unsplash.com/photo-1500534623283-312aade485b7"]', 30),
(1, 'History of Chen Clan Academy', 'The carvings are so detailed. A must-visit for history buffs.', '["https://images.unsplash.com/photo-1500530855697-b586d89ba3ee"]', 8),
(1, 'Walking in Shamian', 'Felt like being in Europe. Very peaceful and photogenic.', '["https://images.unsplash.com/photo-1500534314209-a25ddb2bd429"]', 18);

INSERT INTO favorites (user_id, place_id) VALUES
(1, 1), (1, 4), (2, 3);
