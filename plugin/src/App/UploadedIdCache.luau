--!strict
local HttpService = game:GetService("HttpService")
local FileSystemService = game:GetService("FileSystemService")

local UploadedIdCache = {}

local function getCacheFilePath(assetType: string): string
	return `uploaded_ids_{assetType}.json`
end

-- Load previously uploaded IDs for an asset type
function UploadedIdCache.load(assetType: string): { [number]: boolean }
	local filePath = getCacheFilePath(assetType)
	
	local success, content = pcall(FileSystemService.ReadFile, FileSystemService, filePath)
	if not success or not content then
		return {}
	end
	
	local decoded = pcall(HttpService.JSONDecode, HttpService, content)
	if not decoded then
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
	local filePath = getCacheFilePath(assetType)
	
	-- Load existing IDs
	local success, content = pcall(FileSystemService.ReadFile, FileSystemService, filePath)
	local existingIds: { number } = {}
	
	if success and content then
		local decoded = pcall(HttpService.JSONDecode, HttpService, content)
		if decoded then
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
	
	-- Write back to file
	local encoded = HttpService:JSONEncode(allIds)
	success = pcall(FileSystemService.WriteFile, FileSystemService, filePath, encoded)
	
	return success
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
	local filePath = getCacheFilePath(assetType)
	local success = pcall(FileSystemService.RemoveFile, FileSystemService, filePath)
	return success
end

return UploadedIdCache
