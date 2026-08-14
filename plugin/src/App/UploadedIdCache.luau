--!strict
local HttpService = game:GetService("HttpService")

local UploadedIdCache = {}

local plugin: Plugin = plugin

local function getCacheKey(assetType: string): string
	return `uploaded_ids_{assetType}`
end

-- Load previously uploaded IDs for an asset type
function UploadedIdCache.load(assetType: string): { [number]: boolean }
	local cacheKey = getCacheKey(assetType)
	local cached = plugin:GetSetting(cacheKey)
	
	if not cached then
		return {}
	end
	
	local success, decoded = pcall(HttpService.JSONDecode, HttpService, cached)
	if not success or not decoded then
		return {}
	end
	
	-- Convert array to set (map id -> true)
	local idSet: { [number]: boolean } = {}
	for _, id in (decoded :: any) do
		idSet[id] = true
	end
	
	return idSet
end

-- Save uploaded IDs for an asset type
function UploadedIdCache.save(assetType: string, ids: { number }): boolean
	local cacheKey = getCacheKey(assetType)
	local cached = plugin:GetSetting(cacheKey)
	local existingIds: { number } = {}
	
	if cached then
		local success, decoded = pcall(HttpService.JSONDecode, HttpService, cached)
		if success and decoded then
			existingIds = (decoded :: any) or {}
		end
	end
	
	-- Create a set for fast lookup
	local idSet: { [number]: boolean } = {}
	for _, id in existingIds do
		idSet[id] = true
	end
	
	-- Add new IDs to the set
	for _, id in ids do
		idSet[id] = true
	end
	
	-- Convert set back to array
	local allIds: { number } = {}
	for id in idSet do
		table.insert(allIds, id)
	end
	
	table.sort(allIds)
	
	-- Write back to settings
	local encoded = HttpService:JSONEncode(allIds)
	plugin:SetSetting(cacheKey, encoded)
	
	return true
end

-- Get the IDs that haven't been uploaded yet
function UploadedIdCache.filterNewIds(assetType: string, currentIds: { number }): { number }
	local uploadedIds = UploadedIdCache.load(assetType)
	
	local newIds: { number } = {}
	for _, id in currentIds do
		if not uploadedIds[id] then
			table.insert(newIds, id)
		end
	end
	
	return newIds
end

-- Clear the cache for an asset type
function UploadedIdCache.clear(assetType: string): boolean
	local cacheKey = getCacheKey(assetType)
	plugin:SetSetting(cacheKey, nil)
	return true
end

return UploadedIdCache

