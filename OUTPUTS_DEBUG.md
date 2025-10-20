# 🔍 Debug: Outputs del Remote State

## Problema
El `terraform_remote_state` está funcionando (tiene 14 atributos), pero algunos nombres de outputs no coinciden.

## Errores encontrados:
1. `cluster_oidc_provider_arn` - No existe
2. `eso_role_name` - No existe

## Nombres comunes de outputs en módulos EKS:

### Para OIDC Provider:
- `oidc_provider_arn` ✅ (más común)
- `cluster_oidc_provider_arn`
- `eks_oidc_provider_arn`
- `oidc_provider_arn`

### Para External Secrets Operator:
- `external_secrets_irsa_role_arn` ✅ (más común)
- `eso_irsa_role_arn`
- `external_secrets_role_arn`
- `eso_role_name`

### Para AWS Load Balancer Controller:
- `aws_load_balancer_controller_irsa_role_arn` ✅ (más común)
- `aws_lb_controller_role_arn`
- `alb_controller_role_arn`

## Cómo verificar qué outputs están disponibles:

### Opción 1: Usar el script de debug
```bash
cd infra/terraform/aws
./check_outputs.sh
```

### Opción 2: Usar terraform console
```bash
cd infra/terraform/aws
terraform init
terraform console
# Luego ejecutar:
keys(data.terraform_remote_state.shared.outputs)
```

### Opción 3: Verificar en el repo infrastructure-shared
```bash
cd infrastructure-shared/terraform
terraform output
```

## Solución temporal:
He comentado las líneas problemáticas para que puedas hacer `terraform plan` y ver qué outputs están realmente disponibles.

## Próximos pasos:
1. Ejecutar `terraform plan` para ver qué outputs están disponibles
2. Ajustar los nombres según los outputs reales
3. Descomentar las líneas una vez que tengamos los nombres correctos
