return {
  -- You can specify the plugin using the GitHub repository path
-- copilot {{{
{
	"zbirenbaum/copilot.lua",
	cmd = "Copilot",
	event = "InsertEnter",
	config = function()
		require("copilot").setup({
			panel = {
				enabled = true,
				auto_refresh = true,
			},
			suggestion = {
				enabled = true,
				auto_trigger = true,
			},
		})
	end,
},
-- }}}

}
