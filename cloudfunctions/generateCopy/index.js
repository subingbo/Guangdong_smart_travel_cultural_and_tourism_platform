// 云函数入口文件
const cloud = require('wx-server-sdk')

cloud.init({ env: cloud.DYNAMIC_CURRENT_ENV }) // 使用当前云环境

// 模拟 AI 生成模板库
const TEMPLATES = {
  'en': {
    'instagram': [
      "✨ Experience the authentic taste of Lingnan! 🍜\n\n{text} - A silky, smooth delight that melts in your heart. ❤️\n\n📍 Visit us today!\n#LingnanFood #Authentic #{text} #Foodie #TravelChina",
      "📸 Picture perfect and delicious! 😋\n\nOur {text} is made with love and tradition. Come taste the difference! 👇\n\n#Guangdong #Food #Yum #{text} #MustTry"
    ],
    'twitter': [
      "Craving something special? Try our {text}! � Authentic, fresh, and delicious. #CantoneseFood #{text}",
      "Best {text} in town! Come by and say hi. 👋 #Foodie #Travel #Guangdong"
    ],
    'facebook': [
      "Discover the hidden gem of Lingnan cuisine. 🌸\n\nOur {text} is crafted using traditional methods passed down through generations. It's not just food; it's a piece of history. Perfect for sharing with family and friends. 👨‍👩‍👧‍👦",
      "We are proud to serve the finest {text}. Whether you are a local or a visitor, this is a taste you won't forget. Join us for a memorable meal! 🍽️"
    ]
  },
  'ja': {
    'instagram': [
      "✨ 本場の味を体験しよう！ 🍜\n\n{text} - 口の中でとろける、至福の味わい。❤️\n\n📍 ぜひお越しください！\n#嶺南グルメ #本場の味 #{text} #スイーツ #中国旅行",
      "📸 映える＆美味しい！ 😋\n\n私たちの {text} は伝統と愛情で作られています。ぜひ味わってください！ 👇\n\n#広東 #グルメ #美味しい #{text} #必食"
    ],
    'twitter': [
      "特別な何かをお探しですか？ 当店の {text} をお試しください！ � 本格的で新鮮。 #広東料理 #{text}",
      "街で一番の {text} です！ ぜひお立ち寄りください。 👋 #グルメ #旅行 #広東"
    ],
    'facebook': [
      "嶺南料理の隠れた名品を発見。 🌸\n\n私たちの {text} は、世代を超えて受け継がれてきた伝統的な製法で作られています。それは単なる食べ物ではなく、歴史そのものです。ご家族やご友人とのシェアに最適です。 👨‍👩‍👧‍👦",
      "最高級の {text} をご提供できることを誇りに思います。 地元の方も観光客の方も、忘れられない味になること間違いなし。 思い出に残る食事をぜひ！ 🍽️"
    ]
  },
  'ko': {
    'instagram': [
      "✨ 링난의 진정한 맛을 경험해보세요! 🍜\n\n{text} - 입안에서 사르르 녹는 환상적인 맛. ❤️\n\n📍 오늘 방문해보세요!\n#링난맛집 #진정한맛 #{text} #디저트 #중국여행",
      "📸 사진 찍기 좋고 맛도 최고! 😋\n\n저희 {text}는 정성과 전통으로 만들어집니다. 그 차이를 느껴보세요! 👇\n\n#광동 #맛집 #맛있다 #{text} #필수코스"
    ],
    'twitter': [
      "특별한 맛을 찾고 계신가요? 저희 {text}를 드셔보세요! � 정통의 신선한 맛. #광동요리 #{text}",
      "이 동네 최고의 {text}! 한번 들러주세요. 👋 #맛집 #여행 #광동"
    ],
    'facebook': [
      "링난 요리의 숨겨진 보석을 발견하세요. 🌸\n\n저희 {text}는 대대로 내려오는 전통 방식으로 만들어집니다. 단순한 음식이 아니라 역사의 한 조각입니다. 가족, 친구들과 함께 나누기에 완벽합니다. 👨‍👩‍👧‍👦",
      "최고급 {text}를 선보이게 되어 자랑스럽습니다. 현지인이든 관광객이든 잊지 못할 맛이 될 것입니다. 기억에 남는 식사를 함께하세요! �️"
    ]
  }
}

// 辅助函数：获取随机模板并替换
function generateContent(lang, style, text) {
  const langTemplates = TEMPLATES[lang] || TEMPLATES['en']
  const styleTemplates = langTemplates[style] || langTemplates['instagram']
  
  // 随机选择一个模板
  const randomIndex = Math.floor(Math.random() * styleTemplates.length)
  let content = styleTemplates[randomIndex]
  
  // 替换文本
  const safeText = text || (lang === 'en' ? 'Signature Dish' : (lang === 'ja' ? '看板メニュー' : '시그니처 메뉴'))
  
  // 全局替换 {text}
  content = content.replace(/\{text\}/g, safeText)
  
  return content
}

// 云函数入口函数
exports.main = async (event, context) => {
  const { text, languages = ['en'], style = 'instagram' } = event
  
  // 模拟 AI 处理延迟 (500ms - 1500ms)
  const delay = Math.floor(Math.random() * 1000) + 500
  await new Promise(resolve => setTimeout(resolve, delay))

  const results = {}
  
  for (const lang of languages) {
    results[lang] = generateContent(lang, style, text)
  }

  return {
    success: true,
    data: results
  }
}
