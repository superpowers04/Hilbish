local commander = require 'commander'
local fs = require 'fs'
local lunacolors = require 'lunacolors'

commander.register('doc', function(args)
	local moddocPath = hilbish.dataDir .. '/docs/'
	local apidocHeader = [[
# %s
{grayBg}  {white}{italic}%s  {reset}

]]

	if #args > 0 then
		local mod = args[1]

		local f = io.open(moddocPath .. mod .. '.md', 'rb')
		local funcdocs = nil
		if not f then
			-- assume subdir
			-- dataDir/docs/<mod>/<mod>.md
			moddocPath = moddocPath .. mod .. '/'
			local subdocName = args[2]
			if not subdocName then
				subdocName = 'index'
			end
			f = io.open(moddocPath .. subdocName .. '.md', 'rb')
			if not f then
				print('No documentation found for ' .. mod .. '.')
				return
			end
			funcdocs = f:read '*a'
			local moddocs = table.filter(fs.readdir(moddocPath), function(f) return f ~= 'index.md' end)
			local subdocs = table.map(moddocs, function(fname)
				return lunacolors.underline(lunacolors.blue(string.gsub(fname, '.md', '')))
			end)
			if subdocName == 'index' then
				funcdocs = funcdocs .. '\nSubdocs: ' .. table.concat(subdocs, ', ')
			end
		end

		if not funcdocs then
			funcdocs = f:read '*a'
		end
		local valsStr = funcdocs:match '%-%-%-\n([^%-%-%-]+)\n'
		local vals = {}
		if valsStr then
			local _, endpos = funcdocs:find('---\n' .. valsStr .. '\n---\n\n', 1, true)
			funcdocs = funcdocs:sub(endpos + 1, #funcdocs)

			-- parse vals
			local lines = string.split(valsStr, '\n')
			for _, line in ipairs(lines) do
				local key = line:match '(%w+): '
				local val = line:match '^%w+: (.-)$'

				vals[key] = val
			end
		end
		if mod == 'api' then
			funcdocs = string.format(apidocHeader, vals.name, vals.description) .. funcdocs
		end
		local backtickOccurence = 0
		local formattedFuncs = lunacolors.format(funcdocs:sub(1, #funcdocs - 1):gsub('`', function()
			backtickOccurence = backtickOccurence + 1
			if backtickOccurence % 2 == 0 then
				return '{reset}'
			else
				return '{underline}{green}'
			end
		end):gsub('#+.-\n', function(t)
			return '{bold}{magenta}' .. t .. '{reset}'
		end))
		print(formattedFuncs)
		f:close()

		return
	end
	local modules = table.map(fs.readdir(moddocPath), function(f)
		return lunacolors.underline(lunacolors.blue(string.gsub(f, '.md', '')))
	end)

	io.write [[
Welcome to Hilbish's doc tool! Here you can find documentation for builtin
functions and other things.

Usage: doc <section> [subdoc]
A section is a module or a literal section and a subdoc is a subsection for it.

Available sections: ]]
	io.flush()

	print(table.concat(modules, ', '))
end)