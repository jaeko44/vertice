package opennebula

import (
	"bytes"
	"fmt"
	"github.com/megamsys/megamd/iaas"
	"github.com/megamsys/megamd/global"
	"github.com/megamsys/megamd/provisioner"
	"github.com/tsuru/config"
	"strings"
	"encoding/json"
)

func Init() {
	iaas.RegisterIaasProvider("opennebula", &OpenNebulaIaaS{})
}

type OpenNebulaIaaS struct{}

func (i *OpenNebulaIaaS) DeleteMachine(pdc *global.PredefClouds, assembly *provisioner.AssemblyResult) (string, error) {

	keys, err_keys := iaas.GetAccessKeys(pdc)
     if err_keys != nil {
     	return "", err_keys
     }
     
     str, err := buildDelCommand(iaas.GetPlugins("opennebula"), pdc, "delete")
	if err != nil {
	return "", err
	 }
	//str = str + " -P " + " -y "
	str = str + " -N " + assembly.Name + "." + assembly.Components[0].Inputs.Domain
	str = str + " -A " + keys.AccessKey
	str = str + " -K " + keys.SecretKey

   knifePath, kerr := config.GetString("knife:path")
	if kerr != nil {
		return "", kerr
	}
	str = strings.Replace(str, "-c", "-c "+knifePath, -1)
	str = strings.Replace(str, "<node_name>", assembly.Name + "." + assembly.Components[0].Inputs.Domain, -1 )
   

return str, nil	
}

func (i *OpenNebulaIaaS) CreateMachine(pdc *global.PredefClouds, assembly *provisioner.AssemblyResult) (string, error) {
	keys, err_keys := iaas.GetAccessKeys(pdc)
	if err_keys != nil {
		return "", err_keys
	}

	str, err := buildCommand(iaas.GetPlugins("opennebula"), pdc, "create")
	if err != nil {
		return "", err
	}
	str = str + " -N " + assembly.Name + "." + assembly.Components[0].Inputs.Domain
	str = str + " -A " + keys.AccessKey
	str = str + " -K " + keys.SecretKey
	
	riak, err_riak := config.GetString("api:server")
	if err_riak != nil {
		return "", err_riak
	}
	
	recipe, err_recipe := config.GetString("knife:recipe")
	if err_recipe != nil {
		return "", err_recipe
	}
	
	str = str + " --run-list \"" + "recipe[" + recipe + "]" + "\""
	attributes := &iaas.Attributes{RiakHost: riak, AccountID: pdc.Accounts_id, AssemblyID: assembly.Id}
    b, aerr := json.Marshal(attributes)
    if aerr != nil {
        fmt.Println(aerr)
        return "", aerr
    }
	str = str + " --json-attributes " + string(b)
	
	//strings.Replace(str,"-c","-c "+assembly.Name+"."+assembly.Components[0].Inputs.Domain,-1)
	knifePath, kerr := config.GetString("knife:path")
	if kerr != nil {
		return "", kerr
	}
	str = strings.Replace(str, "-c", "-c "+knifePath, -1)
	return str, nil
}

func buildDelCommand(plugin *iaas.Plugins, pdc *global.PredefClouds, command string) (string, error) {
	var buffer bytes.Buffer
	if len(plugin.Tool) > 0 {
		buffer.WriteString(plugin.Tool)
	} else {
		return "", fmt.Errorf("Plugin tool doesn't loaded")
	}
	if command == "delete" {
		if len(plugin.Command.Delete) > 0 {
			buffer.WriteString(" " + plugin.Command.Delete)
		} else {
			return "", fmt.Errorf("Plugin commands doesn't loaded")
		}
	}
	return buffer.String(), nil 
	
}	

func buildCommand(plugin *iaas.Plugins, pdc *global.PredefClouds, command string) (string, error) {
	var buffer bytes.Buffer
	if len(plugin.Tool) > 0 {
		buffer.WriteString(plugin.Tool)
	} else {
		return "", fmt.Errorf("Plugin tool doesn't loaded")
	}

	if command == "create" {
		if len(plugin.Command.Create) > 0 {
			buffer.WriteString(" " + plugin.Command.Create)
		} else {
			return "", fmt.Errorf("Plugin commands doesn't loaded")
		}
	}

	if len(pdc.Spec.Image) > 0 {
		buffer.WriteString(" --template-name " + pdc.Spec.Image)
	} else {
		return "", fmt.Errorf("Image doesn't loaded")
	}

	if len(pdc.Access.Sshuser) > 0 {
		buffer.WriteString(" -x " + pdc.Access.Sshuser)
	} else {
		return "", fmt.Errorf("Ssh user value doesn't loaded")
	}

	if len(pdc.Access.IdentityFile) > 0 {
		ifile, err := iaas.GetIdentityFileLocation(pdc.Access.IdentityFile)
		if err != nil {
			return "", fmt.Errorf("Identity file doesn't loaded")
		}
		buffer.WriteString(" --identity-file " + ifile + ".key")
	} else {
		return "", fmt.Errorf("Identity file doesn't loaded")
	}

	if len(pdc.Access.Zone) > 0 {
		buffer.WriteString(" --endpoint " + pdc.Access.Zone)
	} else {
		return "", fmt.Errorf("Zone doesn't loaded")
	}

	return buffer.String(), nil
}