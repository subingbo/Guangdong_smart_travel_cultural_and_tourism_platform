$ErrorActionPreference = 'Stop'

function SafeInvoke($scriptBlock) {
  try {
    $r = & $scriptBlock
    return @{ ok = $true; result = $r }
  } catch {
    return @{ ok = $false; status = $_.Exception.Response.StatusCode.value__ ; message = $_.ErrorDetails.Message }
  }
}

function Test-Backend($name, $base) {
  Write-Host "===== $name ($base) ====="

  # 1. public endpoints
  $scenics = Invoke-RestMethod "$base/api/scenics"
  Write-Host "scenics: $($scenics.data.Count) items, success=$($scenics.success)"
  $places = Invoke-RestMethod "$base/api/places?tag=food"
  Write-Host "places(food): $($places.data.Count) items"

  # 2. register + login
  $uname = "tester_$name" + '_' + (Get-Random)
  $reg = Invoke-RestMethod -Method Post "$base/api/auth/register" -ContentType 'application/json' `
    -Body (@{ username = $uname; password = 'test123'; nickname = 'Tester' } | ConvertTo-Json)
  Write-Host "register: success=$($reg.success), hasToken=$([bool]$reg.data.token)"
  $token = $reg.data.token
  $headers = @{ Authorization = "Bearer $token" }

  $login = Invoke-RestMethod -Method Post "$base/api/auth/login" -ContentType 'application/json' `
    -Body (@{ username = $uname; password = 'test123' } | ConvertTo-Json)
  Write-Host "login: success=$($login.success)"

  # 3. me + stats
  $me = Invoke-RestMethod "$base/api/me" -Headers $headers
  Write-Host "me: id=$($me.data.id), nickname=$($me.data.nickname)"
  $stats = Invoke-RestMethod "$base/api/me/stats" -Headers $headers
  Write-Host "stats: fav=$($stats.data.favorites), routes=$($stats.data.routes), posts=$($stats.data.posts)"

  # 4. favorites
  Invoke-RestMethod -Method Post "$base/api/favorites" -Headers $headers -ContentType 'application/json' -Body '{"placeId":2}'
  $favList = Invoke-RestMethod "$base/api/favorites" -Headers $headers
  Write-Host "favorites after add: $($favList.data.Count)"
  $dup = SafeInvoke { Invoke-RestMethod -Method Post "$base/api/favorites" -Headers $headers -ContentType 'application/json' -Body '{"placeId":2}' }
  Write-Host "duplicate add -> ok=$($dup.ok), status=$($dup.status) (expect 409)"
  Invoke-RestMethod -Method Delete "$base/api/favorites/2" -Headers $headers
  $favList2 = Invoke-RestMethod "$base/api/favorites" -Headers $headers
  Write-Host "favorites after delete: $($favList2.data.Count)"

  # 5. blogs: anonymous feed, publish, like
  $feed = Invoke-RestMethod "$base/api/blogs"
  Write-Host "anonymous feed: $($feed.data.Count) posts, likedByMe=$($feed.data[0].likedByMe)"
  $newBlog = Invoke-RestMethod -Method Post "$base/api/blogs" -Headers $headers -ContentType 'application/json' `
    -Body (@{ title = 'Test Post'; content = 'Hello from automated test' } | ConvertTo-Json)
  $blogId = $newBlog.data.id
  Write-Host "publish: blogId=$blogId"
  $liked = Invoke-RestMethod -Method Post "$base/api/blogs/$blogId/like" -Headers $headers
  Write-Host "like -> likes=$($liked.data.likes)"
  $detail = Invoke-RestMethod "$base/api/blogs/$blogId" -Headers $headers
  Write-Host "detail: title=$($detail.data.title)"

  # 6. guide chat without AI key -> 502
  $chat = SafeInvoke { Invoke-RestMethod -Method Post "$base/api/guide/chat" -Headers $headers -ContentType 'application/json' -Body '{"question":"hi"}' }
  Write-Host "guide chat (no key) -> ok=$($chat.ok), status=$($chat.status) (expect 502)"

  # 7. route generate without AI key -> 502
  $route = SafeInvoke { Invoke-RestMethod -Method Post "$base/api/routes/generate" -Headers $headers -ContentType 'application/json' -Body '{"days":2,"interests":["foodie"],"pace":"relaxed","mode":"driving"}' }
  Write-Host "route generate (no key) -> ok=$($route.ok), status=$($route.status) (expect 502)"

  # 8. protected without token -> 401
  $noAuth = SafeInvoke { Invoke-RestMethod "$base/api/me" }
  Write-Host "no token /api/me -> ok=$($noAuth.ok), status=$($noAuth.status) (expect 401)"

  # 9. wrong password -> 401
  $badLogin = SafeInvoke { Invoke-RestMethod -Method Post "$base/api/auth/login" -ContentType 'application/json' -Body (@{ username = $uname; password = 'wrongpass' } | ConvertTo-Json) }
  Write-Host "bad login -> ok=$($badLogin.ok), status=$($badLogin.status) (expect 401)"
}

Test-Backend 'Java' 'http://localhost:8080'
Test-Backend 'Go' 'http://localhost:8081'
