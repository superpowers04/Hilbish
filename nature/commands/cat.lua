local commander = require 'commander'
local fs = require 'fs'

commander.register('cat', function(args, sinks)

	if #args == 0 then
		sinks.out:writeln "\nusage: cat [file]..."
		return 0
	end

	local exit = 0
	for _, fName in ipairs(args) do
		local f = io.open(fName)
		if f == nil then
			exit = 1
			sinks.out:writeln(('cat: %s: no such file or directory'):format(fName))
			goto continue
		end
		local out,err = f:read('*a')
		if out == nil then
			exit = 1
			sinks.out:writeln(('cat: %s: %s'):format(fName,err and err:match(': (.+)') 
				or 'read returned invalid response' ))
			goto continue
		end
		sinks.out:writeln(out)
		::continue::
	end
	io.flush()
	return exit
end)
