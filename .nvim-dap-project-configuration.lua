local dap_go = require('dap-go')
local dap = require('dap')
local dap_ui = require('dapui')

table.insert(dap.configurations.go, {
  type = 'delve',
  name = 'Container debugging (/wd:34567)',
  mode = 'remote',
  request = 'attach',
  substitutePath = {
    { from = '${workspaceFolder}', to = '/wd' },
  },
})

dap.adapters.delve = {
  type = 'server',
  host = 'localhost',
  port = '34567'
}
return {}
