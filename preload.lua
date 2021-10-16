-- The preload file initializes everything else for our shell

local fs = require 'fs'
local commander = require 'commander'
local bait = require 'bait'
local oldDir = hilbish.cwd()

local shlvl = tonumber(os.getenv 'SHLVL')
if shlvl ~= nil then os.setenv('SHLVL', shlvl + 1) else os.setenv('SHLVL', 1) end

-- Builtins
commander.register('cd', function (args)
	if #args > 0 then
		local path = ''
		for i = 1, #args do
			path = path .. tostring(args[i]) .. ' '
		end
		path = path:gsub('$%$','\0'):gsub('${([%w_]+)}', os.getenv)
		:gsub('$([%w_]+)', os.getenv):gsub('%z','$'):gsub('^%s*(.-)%s*$', '%1')

        if path == '-' then
            path = oldDir
            print(path)
        end
        oldDir = hilbish.cwd()

		local ok, err = pcall(function() fs.cd(path) end)
		if not ok then
			if err == 1 then
				print('directory does not exist')
			end
			return err
		end
		bait.throw('cd', path)
		return
	end
	fs.cd(hilbish.home)
	bait.throw('cd', hilbish.home)

	return
end)

commander.register('exit', function()
	os.exit(0)
end)

commander.register('doc', function(args)
	local moddocPath = hilbish.dataDir .. '/docs/'
	local globalDesc = [[
These are the global Hilbish functions that are always available and not part of a module.]]
	if #args > 0 then
		local mod = ''
		for i = 1, #args do
			mod = mod .. tostring(args[i]) .. ' '
		end
		mod = mod:gsub('^%s*(.-)%s*$', '%1')

		local f = io.open(moddocPath .. mod .. '.txt', 'rb')
		if not f then 
			print('Could not find docs for module named ' .. mod .. '.')
			return 1
		end

		local desc = (mod == 'global' and globalDesc or getmetatable(require(mod)).__doc)
		local funcdocs = f:read '*a'
		local backtickOccurence = 0
		print(desc .. '\n\n' .. lunacolors.format(funcdocs:sub(1, #funcdocs - 1):gsub('`', function()
			backtickOccurence = backtickOccurence + 1
			if backtickOccurence % 2 == 0 then
				return '{reset}'
			else
				return '{invert}'
			end
		end)))
		f:close()

		return
	end
	local modules = fs.readdir(moddocPath)

	io.write [[
Welcome to Hilbish's doc tool! Here you can find documentation for builtin
functions and other things.

Usage: doc <module>

Available modules: ]]

	local mods = ''
	for i = 1, #modules do
		mods = mods .. tostring(modules[i]):gsub('.txt', '') .. ', '
	end
	print(mods)

	return
end)

do
	local virt_G = { }

	setmetatable(_G, {
		__index = function (_, key)
			local got_virt = virt_G[key]
			if got_virt ~= nil then
				return got_virt
			end

			virt_G[key] = os.getenv(key)
			return virt_G[key]
		end,

		__newindex = function (_, key, value)
			if type(value) == 'string' then
				os.setenv(key, value)
				virt_G[key] = value
			else
				if type(virt_G[key]) == 'string' then
					os.setenv(key, '')
				end
				virt_G[key] = value
			end
		end,
	})

	bait.catch('command.exit', function ()
		for key, value in pairs(virt_G) do
			if type(value) == 'string' then
				virt_G[key] = os.getenv(key)
			end
		end
	end)
end

-- Function additions to Lua standard library
function string.split(str, delimiter)
	local result = {}
	local from = 1

	local delim_from, delim_to = string.find(str, delimiter, from)

	while delim_from do
		table.insert(result, string.sub(str, from, delim_from - 1))
		from = delim_to + 1
		delim_from, delim_to = string.find(str, delimiter, from)
	end

	table.insert(result, string.sub(str, from))

	return result
end

-- Hook handles
bait.catch('command.not-found', function(cmd)
	print(string.format('hilbish: %s not found', cmd))
end)

