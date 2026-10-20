const cloud = require('wx-server-sdk')

cloud.init({ env: cloud.DYNAMIC_CURRENT_ENV })

const db = cloud.database()
const _ = db.command

const now = () => db.serverDate()

const scenicImages = [
  {
    image_url: 'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee',
    title_cn: '越秀山晨景',
    title_en: 'Morning at Yuexiu Hill',
    location: 'Guangzhou',
    tag: 'nature',
    description_cn: '清晨的越秀山空气清新，适合慢跑与散步。',
    description_en: 'Fresh morning air at Yuexiu Hill, perfect for jogging and strolling.'
  },
  {
    image_url: 'https://images.unsplash.com/photo-1500534314209-a25ddb2bd429',
    title_cn: '珠江夜景',
    title_en: 'Pearl River Night View',
    location: 'Guangzhou',
    tag: 'city',
    description_cn: '珠江两岸灯光璀璨，是广州夜游必看。',
    description_en: 'Brilliant lights along the Pearl River, a must-see night view.'
  },
  {
    image_url: 'https://images.unsplash.com/photo-1500534623283-312aade485b7',
    title_cn: '陈家祠古韵',
    title_en: 'Chen Clan Ancestral Hall',
    location: 'Guangzhou',
    tag: 'culture',
    description_cn: '岭南建筑的代表之作，雕刻精美。',
    description_en: 'A masterpiece of Lingnan architecture with exquisite carvings.'
  },
  {
    image_url: 'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee',
    title_cn: '沙面岛午后',
    title_en: 'Shamian Island Afternoon',
    location: 'Guangzhou',
    tag: 'heritage',
    description_cn: '欧式建筑与骑楼交织，适合拍照打卡。',
    description_en: 'European-style buildings and arcades, great for photos.'
  },
  {
    image_url: 'https://images.unsplash.com/photo-1500534314209-a25ddb2bd429',
    title_cn: '广州塔远眺',
    title_en: 'Canton Tower View',
    location: 'Guangzhou',
    tag: 'landmark',
    description_cn: '城市地标，视野开阔，夜景更迷人。',
    description_en: 'City landmark with expansive views and stunning night lights.'
  }
]

const places = [
  {
    name_cn: '越秀公园',
    name_en: 'Yuexiu Park',
    lat: 23.1387,
    lng: 113.2644,
    tags: ['park', 'history'],
    indoor: false,
    duration: 120,
    description_cn: '广州最大的综合性公园之一。',
    description_en: 'One of the largest comprehensive parks in Guangzhou.'
  },
  {
    name_cn: '陈家祠',
    name_en: 'Chen Clan Ancestral Hall',
    lat: 23.1286,
    lng: 113.2514,
    tags: ['culture', 'architecture'],
    indoor: true,
    duration: 90,
    description_cn: '岭南建筑代表，民间工艺展示。',
    description_en: 'Representative Lingnan architecture with folk arts.'
  },
  {
    name_cn: '沙面岛',
    name_en: 'Shamian Island',
    lat: 23.1146,
    lng: 113.2488,
    tags: ['heritage', 'photo'],
    indoor: false,
    duration: 60,
    description_cn: '欧陆风情街区，适合漫步。',
    description_en: 'European-style district ideal for walking.'
  },
  {
    name_cn: '广州塔',
    name_en: 'Canton Tower',
    lat: 23.1065,
    lng: 113.3247,
    tags: ['landmark', 'view'],
    indoor: false,
    duration: 120,
    description_cn: '广州城市地标，登塔观景。',
    description_en: 'Guangzhou landmark with observation deck.'
  },
  {
    name_cn: '上下九步行街',
    name_en: 'Shangxiajiu Pedestrian Street',
    lat: 23.1259,
    lng: 113.2443,
    tags: ['shopping', 'food'],
    indoor: false,
    duration: 90,
    description_cn: '岭南风情商业街，美食云集。',
    description_en: 'Lingnan-style commercial street with local food.'
  }
]

const knowledge = [
  {
    title_cn: '醒狮文化',
    title_en: 'Lion Dance',
    content_cn: '醒狮是岭南传统民俗，寓意辟邪纳福。',
    content_en: 'Lion dance is a Lingnan folk tradition symbolizing good luck.',
    tags: ['culture', 'festival']
  },
  {
    title_cn: '广绣',
    title_en: 'Guangzhou Embroidery',
    content_cn: '广绣色彩鲜艳、层次丰富，是中国四大名绣之一。',
    content_en: 'Guangzhou embroidery is vivid and layered, one of China’s four famous embroideries.',
    tags: ['craft', 'heritage']
  },
  {
    title_cn: '早茶文化',
    title_en: 'Dim Sum Culture',
    content_cn: '饮早茶是广州人的生活方式，讲究一盅两件。',
    content_en: 'Dim sum is a lifestyle in Guangzhou, known for tea and small dishes.',
    tags: ['food', 'culture']
  },
  {
    title_cn: '岭南建筑',
    title_en: 'Lingnan Architecture',
    content_cn: '岭南建筑注重通风采光，常见骑楼和满洲窗。',
    content_en: 'Lingnan architecture focuses on ventilation and light, featuring arcades and Manchurian windows.',
    tags: ['architecture', 'history']
  },
  {
    title_cn: '粤剧',
    title_en: 'Cantonese Opera',
    content_cn: '粤剧唱腔独特，融合戏曲、音乐、武术。',
    content_en: 'Cantonese opera blends drama, music, and martial arts with distinctive singing.',
    tags: ['performance', 'culture']
  }
]

const chatLogs = [
  {
    merchant_id: 'merchant_demo_1',
    user_id: 'user_demo_1',
    message_list: [
      { from: 'user', text: 'Hello', ts: Date.now() - 100000 },
      { from: 'bot', text: 'Hi, how can I help?', ts: Date.now() - 98000 }
    ]
  },
  {
    merchant_id: 'merchant_demo_2',
    user_id: 'user_demo_2',
    message_list: [
      { from: 'user', text: 'Do you have English menu?', ts: Date.now() - 90000 },
      { from: 'bot', text: 'Yes, we do.', ts: Date.now() - 88000 }
    ]
  },
  {
    merchant_id: 'merchant_demo_3',
    user_id: 'user_demo_3',
    message_list: [
      { from: 'user', text: 'Where is Yuexiu Park?', ts: Date.now() - 80000 },
      { from: 'bot', text: 'It is in Guangzhou city center.', ts: Date.now() - 78000 }
    ]
  },
  {
    merchant_id: 'merchant_demo_4',
    user_id: 'user_demo_4',
    message_list: [
      { from: 'user', text: 'Best time to visit?', ts: Date.now() - 70000 },
      { from: 'bot', text: 'Morning or sunset is great.', ts: Date.now() - 68000 }
    ]
  },
  {
    merchant_id: 'merchant_demo_5',
    user_id: 'user_demo_5',
    message_list: [
      { from: 'user', text: 'Do you accept card?', ts: Date.now() - 60000 },
      { from: 'bot', text: 'Yes, Visa and MasterCard.', ts: Date.now() - 58000 }
    ]
  }
]

const users = [
  { nickName: 'Traveler A', avatarUrl: 'https://i.pravatar.cc/150?img=1' },
  { nickName: 'Traveler B', avatarUrl: 'https://i.pravatar.cc/150?img=2' },
  { nickName: 'Traveler C', avatarUrl: 'https://i.pravatar.cc/150?img=3' },
  { nickName: 'Traveler D', avatarUrl: 'https://i.pravatar.cc/150?img=4' },
  { nickName: 'Traveler E', avatarUrl: 'https://i.pravatar.cc/150?img=5' }
]

const travelBlogs = [
  {
    title: 'A Day in Yuexiu Park',
    content: 'The park is huge and beautiful. I saw the Five Rams Statue.',
    images: ['https://images.unsplash.com/photo-1500530855697-b586d89ba3ee'],
    author: { nickName: 'Traveler A', avatarUrl: 'https://i.pravatar.cc/150?img=1' },
    likes: 12
  },
  {
    title: 'Delicious Dim Sum',
    content: 'Tried chicken feet and shrimp dumplings. Absolutely loved it!',
    images: ['https://images.unsplash.com/photo-1500534314209-a25ddb2bd429'],
    author: { nickName: 'Traveler B', avatarUrl: 'https://i.pravatar.cc/150?img=2' },
    likes: 25
  },
  {
    title: 'Night Cruise on Pearl River',
    content: 'The lights are amazing. Highly recommend taking the boat tour.',
    images: ['https://images.unsplash.com/photo-1500534623283-312aade485b7'],
    author: { nickName: 'Traveler C', avatarUrl: 'https://i.pravatar.cc/150?img=3' },
    likes: 30
  },
  {
    title: 'History of Chen Clan Academy',
    content: 'The carvings are so detailed. A must-visit for history buffs.',
    images: ['https://images.unsplash.com/photo-1500530855697-b586d89ba3ee'],
    author: { nickName: 'Traveler D', avatarUrl: 'https://i.pravatar.cc/150?img=4' },
    likes: 8
  },
  {
    title: 'Walking in Shamian',
    content: 'Felt like being in Europe. Very peaceful and photogenic.',
    images: ['https://images.unsplash.com/photo-1500534314209-a25ddb2bd429'],
    author: { nickName: 'Traveler E', avatarUrl: 'https://i.pravatar.cc/150?img=5' },
    likes: 18
  }
]

const userRoutes = [
  {
    days: 1,
    tags: ['history', 'food'],
    route_nodes: [
      { name: 'Chen Clan Ancestral Hall', time: '09:00' },
      { name: 'Dim Sum Lunch', time: '12:00' },
      { name: 'Yuexiu Park', time: '14:00' }
    ]
  },
  {
    days: 2,
    tags: ['city', 'shopping'],
    route_nodes: [
      { name: 'Canton Tower', time: '10:00' },
      { name: 'Zhujiang New Town', time: '13:00' },
      { name: 'Pearl River Night Cruise', time: '20:00' }
    ]
  },
  {
    days: 1,
    tags: ['photo', 'relax'],
    route_nodes: [
      { name: 'Shamian Island', time: '10:00' },
      { name: 'Sacred Heart Cathedral', time: '14:00' },
      { name: 'Beijing Road', time: '17:00' }
    ]
  },
  {
    days: 3,
    tags: ['comprehensive'],
    route_nodes: [
      { name: 'Baiyun Mountain', time: '09:00' },
      { name: 'Sun Yat-sen Memorial Hall', time: '14:00' },
      { name: 'Museum of Western Han Tomb', time: '16:00' }
    ]
  },
  {
    days: 1,
    tags: ['foodie'],
    route_nodes: [
      { name: 'Liwan Lake Park', time: '08:00' },
      { name: 'Panxi Restaurant', time: '10:00' },
      { name: 'Shangxiajiu Street', time: '13:00' }
    ]
  }
]

const favorites = [
  { type: 'place', name: 'Yuexiu Park', tags: ['park'] },
  { type: 'place', name: 'Canton Tower', tags: ['landmark'] },
  { type: 'blog', title: 'Delicious Dim Sum', tags: ['food'] },
  { type: 'route', name: 'Guangzhou 1 Day Trip', tags: ['history'] },
  { type: 'place', name: 'Shamian Island', tags: ['photo'] }
]

const attachMeta = (list, openid) =>
  list.map((item) => {
    const meta = {
      ...item,
      createdAt: now(),
      updatedAt: now()
    }
    // Inject _openid if provided (to make data belong to the caller)
    if (openid) {
      meta._openid = openid
    }
    return meta
  })

const batchInsert = async (collection, data) => {
  const results = []
  for (const item of data) {
    const res = await db.collection(collection).add({ data: item })
    results.push(res._id)
  }
  return results
}

exports.main = async (event, context) => {
  const { OPENID } = cloud.getWXContext()
  const results = {}

  // 1. Public Data (No OpenID injection)
  try {
    results.users = await batchInsert('users', attachMeta(users))
  } catch (e) {
    results.users = e.message
  }

  try {
    results.scenic_images = await batchInsert('scenic_images', attachMeta(scenicImages))
  } catch (e) {
    results.scenic_images = e.message
  }
  
  try {
    results.places = await batchInsert('places', attachMeta(places))
  } catch (e) {
    results.places = e.message
  }
  
  try {
    results.knowledge = await batchInsert('knowledge', attachMeta(knowledge))
  } catch (e) {
    results.knowledge = e.message
  }
  
  try {
    results.chat_logs = await batchInsert('chat_logs', attachMeta(chatLogs))
  } catch (e) {
    results.chat_logs = e.message
  }
  
  // 2. User-specific Data (Inject OpenID if available)
  // These collections are usually user-generated content
  
  try {
    results.travel_blogs = await batchInsert('travel_blogs', attachMeta(travelBlogs, OPENID))
  } catch (e) {
    results.travel_blogs = e.message
  }
  
  try {
    results.user_routes = await batchInsert('user_routes', attachMeta(userRoutes, OPENID))
  } catch (e) {
    results.user_routes = e.message
  }

  try {
    results.favorites = await batchInsert('favorites', attachMeta(favorites, OPENID))
  } catch (e) {
    results.favorites = e.message
  }

  return {
    success: true,
    results
  }
}
