const cloud = require('wx-server-sdk')

cloud.init({
  env: cloud.DYNAMIC_CURRENT_ENV
})

const db = cloud.database()

exports.main = async (event, context) => {
  const collections = [
    'users',
    'scenic_images',
    'places',
    'knowledge',
    'chat_logs',
    'travel_blogs',
  'user_routes',
  'favorites'
]
  
  const results = {}
  
  for (const name of collections) {
    try {
      await db.createCollection(name)
      results[name] = 'Created'
    } catch (e) {
      if (e.errMsg.includes('Collection already exists') || e.errMsg.includes('Table exist')) {
        results[name] = 'Already Exists'
      } else {
        results[name] = `Error: ${e.errMsg}`
      }
    }
  }
  
  return {
    success: true,
    results
  }
}
