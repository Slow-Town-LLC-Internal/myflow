
vim.cmd([[
  set runtimepath^=~/.vim runtimepath+=~/.vim/after
  let &packpath = &runtimepath
  source ~/.vimrc
]])
require("config.lazy")
require("lazy").setup("plugins")
vim.g.python3_host_prog = '~/adminvenv/bin/python'
vim.opt.laststatus = 3

require'nvim-treesitter.configs'.setup {
   ensure_installed = { "terraform" }, -- List of languages to install automatically
   highlight = {
     enable = true,              -- Enable syntax highlighting
     additional_vim_regex_highlighting = false, -- Do NOT enable additional vim highlighting
   },
}

vim.cmd("colorscheme habamax")

