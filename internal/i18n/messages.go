package i18n

// messages contains all translations keyed by language, then by message key.
var messages = map[string]map[string]string{
	// ═══════════════════════════════════════════════════
	// Portuguese (Brazil)
	// ═══════════════════════════════════════════════════
	LangPtBR: {
		// --- Auth / Credentials ---
		"invalid_credentials":            "Credenciais inválidas",
		"account_not_active":             "Conta não está ativa",
		"account_suspended":              "Conta suspensa",
		"authorization_required":         "Autorização necessária",
		"token_invalidated":              "Token invalidado",
		"invalid_token":                  "Token inválido",
		"token_required":                 "Token é obrigatório",
		"email_already_in_use":           "E-mail já está em uso",
		"email_already_registered":       "E-mail já cadastrado",
		"email_already_verified":         "E-mail já verificado",
		"invalid_verification_token":     "Token de verificação inválido ou expirado",
		"email_verified":                 "E-mail verificado",
		"verification_email_sent":        "E-mail de verificação enviado",
		"logged_out":                     "Desconectado",
		"user_not_found":                 "Usuário não encontrado",
		"password_changed":               "Senha alterada",
		"current_password_incorrect":     "Senha atual incorreta",
		"failed_change_password":         "Falha ao alterar senha",
		"password_reset_sent":            "Se o e-mail existir, um link de redefinição foi enviado",
		"password_reset_success":         "Senha redefinida com sucesso",
		"password_reset_not_implemented": "Redefinição de senha via token ainda não implementada",

		// --- Tenant ---
		"tenant_not_found":      "Tenant não encontrado",
		"tenant_not_active":     "Tenant não está ativo",
		"not_tenant_member":     "Não é membro deste tenant",
		"failed_generate_token": "Falha ao gerar token",
		"access_denied":         "Acesso negado",
		"url_code_required":     "Código URL é obrigatório",

		// --- Plans ---
		"failed_list_plans":        "Falha ao listar planos",
		"invalid_or_inactive_plan": "Plano inválido ou inativo",

		// --- Promotions ---
		"invalid_promo_code": "Código promocional inválido",

		// --- Subscription ---
		"company_name_required": "Nome da empresa é obrigatório quando é empresa",
		"no_active_plan":        "Nenhum plano ativo",
		"max_users_reached":     "Tenant atingiu o limite de usuários (%d/%d)",
		"user_already_member":   "Usuário já é membro deste tenant",
		"role_not_found_tenant": "Função '%s' não encontrada para este tenant",

		// --- Profile ---
		"profile_not_found":         "Perfil não encontrado",
		"profile_updated":           "Perfil atualizado",
		"failed_update_profile":     "Falha ao atualizar perfil",
		"tenant_profile_updated":    "Perfil do tenant atualizado",
		"only_owner_update_profile": "Apenas o proprietário pode atualizar o perfil do tenant",
		"only_owner_upload_logo":    "Apenas o proprietário pode enviar o logo",

		// --- Upload ---
		"no_file_provided": "Nenhum arquivo fornecido",
		"failed_upload":    "Falha no upload",

		// --- Members ---
		"failed_list_members":      "Falha ao listar membros",
		"member_not_found":         "Membro não encontrado",
		"member_added":             "Membro adicionado",
		"member_role_updated":      "Função do membro atualizada",
		"failed_update_role":       "Falha ao atualizar função",
		"only_owner_remove":        "Apenas o proprietário pode remover membros",
		"cannot_remove_self":       "Não é possível remover a si mesmo",
		"member_removed":           "Membro removido",
		"failed_remove_member":     "Falha ao remover membro",
		"insufficient_permissions": "Permissões insuficientes",

		// --- Roles ---
		"failed_list_roles":       "Falha ao listar funções",
		"role_not_found":          "Função não encontrada",
		"only_owner_manage_roles": "Apenas o proprietário pode gerenciar funções",
		"failed_create_role":      "Falha ao criar função",
		"role_updated":            "Função atualizada",
		"role_deleted":            "Função excluída",

		// --- Permissions ---
		"failed_list_permissions":  "Falha ao listar permissões",
		"only_owner_manage_perms":  "Apenas o proprietário pode gerenciar permissões",
		"permission_assigned":      "Permissão atribuída",
		"permission_removed":       "Permissão removida",
		"failed_assign_permission": "Falha ao atribuir permissão",
		"failed_remove_permission": "Falha ao remover permissão",

		// --- Features ---
		"feature_not_in_plan":    "Recurso '%s' não disponível no seu plano",
		"failed_list_features":   "Falha ao listar recursos",
		"feature_not_found":      "Recurso não encontrado",
		"feature_already_exists": "Recurso já existe",
		"failed_create_feature":  "Falha ao criar recurso",
		"failed_update_feature":  "Falha ao atualizar recurso",
		"failed_delete_feature":  "Falha ao excluir recurso",
		"failed_add_feature":     "Falha ao adicionar recurso",
		"failed_remove_feature":  "Falha ao remover recurso",

		// --- Products ---
		"failed_list_products":  "Falha ao listar produtos",
		"product_not_found":     "Produto não encontrado",
		"failed_create_product": "Falha ao criar produto",
		"product_updated":       "Produto atualizado",
		"failed_update_product": "Falha ao atualizar produto",
		"product_deleted":       "Produto excluído",
		"failed_delete_product": "Falha ao excluir produto",
		"failed_save_image":     "Falha ao salvar registro de imagem",

		// --- Services ---
		"failed_list_services":  "Falha ao listar serviços",
		"service_not_found":     "Serviço não encontrado",
		"failed_create_service": "Falha ao criar serviço",
		"service_updated":       "Serviço atualizado",
		"failed_update_service": "Falha ao atualizar serviço",
		"service_deleted":       "Serviço excluído",
		"failed_delete_service": "Falha ao excluir serviço",

		// --- Settings ---
		"layout_settings_saved": "Configurações de layout salvas",
		"settings_saved":        "Configurações salvas",
		"failed_save_layout":    "Falha ao salvar configurações de layout",
		"failed_save_settings":  "Falha ao salvar configurações",
		"language_updated":      "Idioma atualizado",
		"invalid_language":      "Idioma inválido. Use: pt-BR, pt, en, es",

		// --- App Users ---
		"failed_list_app_users":  "Falha ao listar app users",
		"app_user_not_found":     "App user não encontrado",
		"status_updated":         "Status atualizado",
		"failed_update_status":   "Falha ao atualizar status",
		"app_user_deleted":       "App user excluído",
		"failed_delete_app_user": "Falha ao excluir app user",

		// --- Admin specific ---
		"admin_not_found":          "Administrador não encontrado",
		"permission_denied":        "Permissão negada",
		"email_already_exists":     "E-mail já existe",
		"failed_list_admins":       "Falha ao listar administradores",
		"failed_update":            "Falha ao atualizar",
		"failed_delete":            "Falha ao excluir",
		"failed_update_password":   "Falha ao atualizar senha",
		"invalid_current_password": "Senha atual inválida",

		// --- Admin Roles ---
		"role_already_exists": "Função já existe",
		"failed_assign_role":  "Falha ao atribuir função",
		"failed_remove_role":  "Falha ao remover função",
		"failed_delete_role":  "Falha ao excluir função",

		// --- Admin Tenants ---
		"failed_list_tenants":        "Falha ao listar tenants",
		"tenant_already_exists":      "Tenant já existe",
		"plan_not_found":             "Plano não encontrado",
		"transaction_failed":         "Falha na transação",
		"failed_create_plan":         "Falha ao criar plano",
		"owner_creation_failed":      "Falha ao criar proprietário: %s",
		"commit_failed":              "Falha ao confirmar transação",
		"failed_update_tenant":       "Falha ao atualizar tenant",
		"failed_delete_tenant":       "Falha ao excluir tenant",
		"failed_update_status_admin": "Falha ao atualizar status",
		"failed_get_history":         "Falha ao obter histórico",
		"failed_list_members_admin":  "Falha ao listar membros",

		// --- Admin Plans ---
		"failed_list_plans_admin":  "Falha ao listar planos",
		"failed_create_plan_admin": "Falha ao criar plano",
		"plan_not_found_admin":     "Plano não encontrado",
		"failed_update_plan":       "Falha ao atualizar plano",
		"failed_delete_plan":       "Falha ao excluir plano",

		// --- Admin Promotions ---
		"failed_list_promotions":      "Falha ao listar promoções",
		"failed_create_promotion":     "Falha ao criar promoção",
		"promotion_not_found":         "Promoção não encontrada",
		"failed_update_promotion":     "Falha ao atualizar promoção",
		"failed_deactivate_promotion": "Falha ao desativar promoção",

		// --- Validation templates ---
		"validation.required": "%s é obrigatório",
		"validation.email":    "%s deve ser um e-mail válido",
		"validation.min":      "%s deve ter pelo menos %s caracteres",
		"validation.max":      "%s deve ter no máximo %s caracteres",
		"validation.url":      "%s deve ser uma URL válida",
		"validation.oneof":    "%s deve ser um dos: %s",
		"validation.uuid":     "%s deve ser um UUID válido",
		"validation.gte":      "%s deve ser maior ou igual a %s",
		"validation.lte":      "%s deve ser menor ou igual a %s",
		"validation.len":      "%s deve ter exatamente %s caracteres",
		"validation.default":  "%s é inválido",
	},

	// ═══════════════════════════════════════════════════
	// Portuguese (Portugal)
	// ═══════════════════════════════════════════════════
	LangPt: {
		// --- Auth / Credentials ---
		"invalid_credentials":            "Credenciais inválidas",
		"account_not_active":             "Conta não está ativa",
		"account_suspended":              "Conta suspensa",
		"authorization_required":         "Autorização necessária",
		"token_invalidated":              "Token invalidado",
		"invalid_token":                  "Token inválido",
		"token_required":                 "Token é obrigatório",
		"email_already_in_use":           "E-mail já está em uso",
		"email_already_registered":       "E-mail já registado",
		"email_already_verified":         "E-mail já verificado",
		"invalid_verification_token":     "Token de verificação inválido ou expirado",
		"email_verified":                 "E-mail verificado",
		"verification_email_sent":        "E-mail de verificação enviado",
		"logged_out":                     "Sessão terminada",
		"user_not_found":                 "Utilizador não encontrado",
		"password_changed":               "Palavra-passe alterada",
		"current_password_incorrect":     "Palavra-passe atual incorreta",
		"failed_change_password":         "Falha ao alterar palavra-passe",
		"password_reset_sent":            "Se o e-mail existir, um link de redefinição foi enviado",
		"password_reset_success":         "Palavra-passe redefinida com sucesso",
		"password_reset_not_implemented": "Redefinição de palavra-passe via token ainda não implementada",

		// --- Tenant ---
		"tenant_not_found":      "Tenant não encontrado",
		"tenant_not_active":     "Tenant não está ativo",
		"not_tenant_member":     "Não é membro deste tenant",
		"failed_generate_token": "Falha ao gerar token",
		"access_denied":         "Acesso negado",
		"url_code_required":     "Código URL é obrigatório",

		// --- Plans ---
		"failed_list_plans":        "Falha ao listar planos",
		"invalid_or_inactive_plan": "Plano inválido ou inativo",

		// --- Promotions ---
		"invalid_promo_code": "Código promocional inválido",

		// --- Subscription ---
		"company_name_required": "Nome da empresa é obrigatório quando é empresa",
		"no_active_plan":        "Nenhum plano ativo",
		"max_users_reached":     "Tenant atingiu o limite de utilizadores (%d/%d)",
		"user_already_member":   "Utilizador já é membro deste tenant",
		"role_not_found_tenant": "Função '%s' não encontrada para este tenant",

		// --- Profile ---
		"profile_not_found":         "Perfil não encontrado",
		"profile_updated":           "Perfil atualizado",
		"failed_update_profile":     "Falha ao atualizar perfil",
		"tenant_profile_updated":    "Perfil do tenant atualizado",
		"only_owner_update_profile": "Apenas o proprietário pode atualizar o perfil do tenant",
		"only_owner_upload_logo":    "Apenas o proprietário pode enviar o logótipo",

		// --- Upload ---
		"no_file_provided": "Nenhum ficheiro fornecido",
		"failed_upload":    "Falha no envio",

		// --- Members ---
		"failed_list_members":      "Falha ao listar membros",
		"member_not_found":         "Membro não encontrado",
		"member_added":             "Membro adicionado",
		"member_role_updated":      "Função do membro atualizada",
		"failed_update_role":       "Falha ao atualizar função",
		"only_owner_remove":        "Apenas o proprietário pode remover membros",
		"cannot_remove_self":       "Não é possível remover-se a si próprio",
		"member_removed":           "Membro removido",
		"failed_remove_member":     "Falha ao remover membro",
		"insufficient_permissions": "Permissões insuficientes",

		// --- Roles ---
		"failed_list_roles":       "Falha ao listar funções",
		"role_not_found":          "Função não encontrada",
		"only_owner_manage_roles": "Apenas o proprietário pode gerir funções",
		"failed_create_role":      "Falha ao criar função",
		"role_updated":            "Função atualizada",
		"role_deleted":            "Função eliminada",

		// --- Permissions ---
		"failed_list_permissions":  "Falha ao listar permissões",
		"only_owner_manage_perms":  "Apenas o proprietário pode gerir permissões",
		"permission_assigned":      "Permissão atribuída",
		"permission_removed":       "Permissão removida",
		"failed_assign_permission": "Falha ao atribuir permissão",
		"failed_remove_permission": "Falha ao remover permissão",

		// --- Features ---
		"feature_not_in_plan":    "Recurso '%s' não disponível no seu plano",
		"failed_list_features":   "Falha ao listar recursos",
		"feature_not_found":      "Recurso não encontrado",
		"feature_already_exists": "Recurso já existe",
		"failed_create_feature":  "Falha ao criar recurso",
		"failed_update_feature":  "Falha ao atualizar recurso",
		"failed_delete_feature":  "Falha ao eliminar recurso",
		"failed_add_feature":     "Falha ao adicionar recurso",
		"failed_remove_feature":  "Falha ao remover recurso",

		// --- Products ---
		"failed_list_products":  "Falha ao listar produtos",
		"product_not_found":     "Produto não encontrado",
		"failed_create_product": "Falha ao criar produto",
		"product_updated":       "Produto atualizado",
		"failed_update_product": "Falha ao atualizar produto",
		"product_deleted":       "Produto eliminado",
		"failed_delete_product": "Falha ao eliminar produto",
		"failed_save_image":     "Falha ao guardar registo de imagem",

		// --- Services ---
		"failed_list_services":  "Falha ao listar serviços",
		"service_not_found":     "Serviço não encontrado",
		"failed_create_service": "Falha ao criar serviço",
		"service_updated":       "Serviço atualizado",
		"failed_update_service": "Falha ao atualizar serviço",
		"service_deleted":       "Serviço eliminado",
		"failed_delete_service": "Falha ao eliminar serviço",

		// --- Settings ---
		"layout_settings_saved": "Configurações de layout guardadas",
		"settings_saved":        "Configurações guardadas",
		"failed_save_layout":    "Falha ao guardar configurações de layout",
		"failed_save_settings":  "Falha ao guardar configurações",
		"language_updated":      "Idioma atualizado",
		"invalid_language":      "Idioma inválido. Use: pt-BR, pt, en, es",

		// --- App Users ---
		"failed_list_app_users":  "Falha ao listar app users",
		"app_user_not_found":     "App user não encontrado",
		"status_updated":         "Estado atualizado",
		"failed_update_status":   "Falha ao atualizar estado",
		"app_user_deleted":       "App user eliminado",
		"failed_delete_app_user": "Falha ao eliminar app user",

		// --- Admin specific ---
		"admin_not_found":          "Administrador não encontrado",
		"permission_denied":        "Permissão negada",
		"email_already_exists":     "E-mail já existe",
		"failed_list_admins":       "Falha ao listar administradores",
		"failed_update":            "Falha ao atualizar",
		"failed_delete":            "Falha ao eliminar",
		"failed_update_password":   "Falha ao atualizar palavra-passe",
		"invalid_current_password": "Palavra-passe atual inválida",

		// --- Admin Roles ---
		"role_already_exists": "Função já existe",
		"failed_assign_role":  "Falha ao atribuir função",
		"failed_remove_role":  "Falha ao remover função",
		"failed_delete_role":  "Falha ao eliminar função",

		// --- Admin Tenants ---
		"failed_list_tenants":        "Falha ao listar tenants",
		"tenant_already_exists":      "Tenant já existe",
		"plan_not_found":             "Plano não encontrado",
		"transaction_failed":         "Falha na transação",
		"failed_create_plan":         "Falha ao criar plano",
		"owner_creation_failed":      "Falha ao criar proprietário: %s",
		"commit_failed":              "Falha ao confirmar transação",
		"failed_update_tenant":       "Falha ao atualizar tenant",
		"failed_delete_tenant":       "Falha ao eliminar tenant",
		"failed_update_status_admin": "Falha ao atualizar estado",
		"failed_get_history":         "Falha ao obter histórico",
		"failed_list_members_admin":  "Falha ao listar membros",

		// --- Admin Plans ---
		"failed_list_plans_admin":  "Falha ao listar planos",
		"failed_create_plan_admin": "Falha ao criar plano",
		"plan_not_found_admin":     "Plano não encontrado",
		"failed_update_plan":       "Falha ao atualizar plano",
		"failed_delete_plan":       "Falha ao eliminar plano",

		// --- Admin Promotions ---
		"failed_list_promotions":      "Falha ao listar promoções",
		"failed_create_promotion":     "Falha ao criar promoção",
		"promotion_not_found":         "Promoção não encontrada",
		"failed_update_promotion":     "Falha ao atualizar promoção",
		"failed_deactivate_promotion": "Falha ao desativar promoção",

		// --- Validation templates ---
		"validation.required": "%s é obrigatório",
		"validation.email":    "%s deve ser um e-mail válido",
		"validation.min":      "%s deve ter pelo menos %s caracteres",
		"validation.max":      "%s deve ter no máximo %s caracteres",
		"validation.url":      "%s deve ser um URL válido",
		"validation.oneof":    "%s deve ser um dos: %s",
		"validation.uuid":     "%s deve ser um UUID válido",
		"validation.gte":      "%s deve ser superior ou igual a %s",
		"validation.lte":      "%s deve ser inferior ou igual a %s",
		"validation.len":      "%s deve ter exatamente %s caracteres",
		"validation.default":  "%s é inválido",
	},

	// ═══════════════════════════════════════════════════
	// English
	// ═══════════════════════════════════════════════════
	LangEn: {
		// --- Auth / Credentials ---
		"invalid_credentials":            "Invalid credentials",
		"account_not_active":             "Account is not active",
		"account_suspended":              "Account suspended",
		"authorization_required":         "Authorization required",
		"token_invalidated":              "Token invalidated",
		"invalid_token":                  "Invalid token",
		"token_required":                 "Token is required",
		"email_already_in_use":           "Email already in use",
		"email_already_registered":       "Email already registered",
		"email_already_verified":         "Email already verified",
		"invalid_verification_token":     "Invalid or expired verification token",
		"email_verified":                 "Email verified",
		"verification_email_sent":        "Verification email sent",
		"logged_out":                     "Logged out",
		"user_not_found":                 "User not found",
		"password_changed":               "Password changed",
		"current_password_incorrect":     "Current password is incorrect",
		"failed_change_password":         "Failed to change password",
		"password_reset_sent":            "If the email exists, a reset link was sent",
		"password_reset_success":         "Password reset successfully",
		"password_reset_not_implemented": "Password reset via token not yet implemented",

		// --- Tenant ---
		"tenant_not_found":      "Tenant not found",
		"tenant_not_active":     "Tenant is not active",
		"not_tenant_member":     "Not a member of this tenant",
		"failed_generate_token": "Failed to generate token",
		"access_denied":         "Access denied",
		"url_code_required":     "URL code is required",

		// --- Plans ---
		"failed_list_plans":        "Failed to list plans",
		"invalid_or_inactive_plan": "Invalid or inactive plan",

		// --- Promotions ---
		"invalid_promo_code": "Invalid promotion code",

		// --- Subscription ---
		"company_name_required": "Company name is required when is_company is true",
		"no_active_plan":        "No active plan",
		"max_users_reached":     "Tenant has reached max users (%d/%d)",
		"user_already_member":   "User is already a member of this tenant",
		"role_not_found_tenant": "Role '%s' not found for this tenant",

		// --- Profile ---
		"profile_not_found":         "Profile not found",
		"profile_updated":           "Profile updated",
		"failed_update_profile":     "Failed to update profile",
		"tenant_profile_updated":    "Tenant profile updated",
		"only_owner_update_profile": "Only owner can update tenant profile",
		"only_owner_upload_logo":    "Only owner can upload logo",

		// --- Upload ---
		"no_file_provided": "No file provided",
		"failed_upload":    "Failed to upload",

		// --- Members ---
		"failed_list_members":      "Failed to list members",
		"member_not_found":         "Member not found",
		"member_added":             "Member added",
		"member_role_updated":      "Member role updated",
		"failed_update_role":       "Failed to update role",
		"only_owner_remove":        "Only owner can remove members",
		"cannot_remove_self":       "Cannot remove yourself",
		"member_removed":           "Member removed",
		"failed_remove_member":     "Failed to remove member",
		"insufficient_permissions": "Insufficient permissions",

		// --- Roles ---
		"failed_list_roles":       "Failed to list roles",
		"role_not_found":          "Role not found",
		"only_owner_manage_roles": "Only owner can manage roles",
		"failed_create_role":      "Failed to create role",
		"role_updated":            "Role updated",
		"role_deleted":            "Role deleted",

		// --- Permissions ---
		"failed_list_permissions":  "Failed to list permissions",
		"only_owner_manage_perms":  "Only owner can manage permissions",
		"permission_assigned":      "Permission assigned",
		"permission_removed":       "Permission removed",
		"failed_assign_permission": "Failed to assign permission",
		"failed_remove_permission": "Failed to remove permission",

		// --- Features ---
		"feature_not_in_plan":    "Feature '%s' not available in your plan",
		"failed_list_features":   "Failed to list features",
		"feature_not_found":      "Feature not found",
		"feature_already_exists": "Feature already exists",
		"failed_create_feature":  "Failed to create feature",
		"failed_update_feature":  "Failed to update feature",
		"failed_delete_feature":  "Failed to delete feature",
		"failed_add_feature":     "Failed to add feature",
		"failed_remove_feature":  "Failed to remove feature",

		// --- Products ---
		"failed_list_products":  "Failed to list products",
		"product_not_found":     "Product not found",
		"failed_create_product": "Failed to create product",
		"product_updated":       "Product updated",
		"failed_update_product": "Failed to update product",
		"product_deleted":       "Product deleted",
		"failed_delete_product": "Failed to delete product",
		"failed_save_image":     "Failed to save image record",

		// --- Services ---
		"failed_list_services":  "Failed to list services",
		"service_not_found":     "Service not found",
		"failed_create_service": "Failed to create service",
		"service_updated":       "Service updated",
		"failed_update_service": "Failed to update service",
		"service_deleted":       "Service deleted",
		"failed_delete_service": "Failed to delete service",

		// --- Settings ---
		"layout_settings_saved": "Layout settings saved",
		"settings_saved":        "Settings saved",
		"failed_save_layout":    "Failed to save layout settings",
		"failed_save_settings":  "Failed to save settings",
		"language_updated":      "Language updated",
		"invalid_language":      "Invalid language. Use: pt-BR, pt, en, es",

		// --- App Users ---
		"failed_list_app_users":  "Failed to list app users",
		"app_user_not_found":     "App user not found",
		"status_updated":         "Status updated",
		"failed_update_status":   "Failed to update status",
		"app_user_deleted":       "App user deleted",
		"failed_delete_app_user": "Failed to delete app user",

		// --- Admin specific ---
		"admin_not_found":          "Admin not found",
		"permission_denied":        "Permission denied",
		"email_already_exists":     "Email already exists",
		"failed_list_admins":       "Failed to list admins",
		"failed_update":            "Failed to update",
		"failed_delete":            "Failed to delete",
		"failed_update_password":   "Failed to update password",
		"invalid_current_password": "Invalid current password",

		// --- Admin Roles ---
		"role_already_exists": "Role already exists",
		"failed_assign_role":  "Failed to assign role",
		"failed_remove_role":  "Failed to remove role",
		"failed_delete_role":  "Failed to delete role",

		// --- Admin Tenants ---
		"failed_list_tenants":        "Failed to list tenants",
		"tenant_already_exists":      "Tenant already exists",
		"plan_not_found":             "Plan not found",
		"transaction_failed":         "Transaction failed",
		"failed_create_plan":         "Failed to create plan",
		"owner_creation_failed":      "Owner creation failed: %s",
		"commit_failed":              "Commit failed",
		"failed_update_tenant":       "Failed to update tenant",
		"failed_delete_tenant":       "Failed to delete tenant",
		"failed_update_status_admin": "Failed to update status",
		"failed_get_history":         "Failed to get history",
		"failed_list_members_admin":  "Failed to list members",

		// --- Admin Plans ---
		"failed_list_plans_admin":  "Failed to list plans",
		"failed_create_plan_admin": "Failed to create plan",
		"plan_not_found_admin":     "Plan not found",
		"failed_update_plan":       "Failed to update plan",
		"failed_delete_plan":       "Failed to delete plan",

		// --- Admin Promotions ---
		"failed_list_promotions":      "Failed to list promotions",
		"failed_create_promotion":     "Failed to create promotion",
		"promotion_not_found":         "Promotion not found",
		"failed_update_promotion":     "Failed to update promotion",
		"failed_deactivate_promotion": "Failed to deactivate promotion",

		// --- Validation templates ---
		"validation.required": "%s is required",
		"validation.email":    "%s must be a valid email address",
		"validation.min":      "%s must be at least %s characters",
		"validation.max":      "%s must be at most %s characters",
		"validation.url":      "%s must be a valid URL",
		"validation.oneof":    "%s must be one of: %s",
		"validation.uuid":     "%s must be a valid UUID",
		"validation.gte":      "%s must be greater than or equal to %s",
		"validation.lte":      "%s must be less than or equal to %s",
		"validation.len":      "%s must be exactly %s characters",
		"validation.default":  "%s is invalid",
	},

	// ═══════════════════════════════════════════════════
	// Spanish
	// ═══════════════════════════════════════════════════
	LangEs: {
		// --- Auth / Credentials ---
		"invalid_credentials":            "Credenciales inválidas",
		"account_not_active":             "La cuenta no está activa",
		"account_suspended":              "Cuenta suspendida",
		"authorization_required":         "Autorización requerida",
		"token_invalidated":              "Token invalidado",
		"invalid_token":                  "Token inválido",
		"token_required":                 "Token es obligatorio",
		"email_already_in_use":           "El correo electrónico ya está en uso",
		"email_already_registered":       "El correo electrónico ya está registrado",
		"email_already_verified":         "El correo electrónico ya está verificado",
		"invalid_verification_token":     "Token de verificación inválido o expirado",
		"email_verified":                 "Correo electrónico verificado",
		"verification_email_sent":        "Correo de verificación enviado",
		"logged_out":                     "Sesión cerrada",
		"user_not_found":                 "Usuario no encontrado",
		"password_changed":               "Contraseña cambiada",
		"current_password_incorrect":     "La contraseña actual es incorrecta",
		"failed_change_password":         "Error al cambiar la contraseña",
		"password_reset_sent":            "Si el correo existe, se envió un enlace de restablecimiento",
		"password_reset_success":         "Contraseña restablecida con éxito",
		"password_reset_not_implemented": "Restablecimiento de contraseña por token aún no implementado",

		// --- Tenant ---
		"tenant_not_found":      "Tenant no encontrado",
		"tenant_not_active":     "Tenant no está activo",
		"not_tenant_member":     "No es miembro de este tenant",
		"failed_generate_token": "Error al generar token",
		"access_denied":         "Acceso denegado",
		"url_code_required":     "El código URL es obligatorio",

		// --- Plans ---
		"failed_list_plans":        "Error al listar planes",
		"invalid_or_inactive_plan": "Plan inválido o inactivo",

		// --- Promotions ---
		"invalid_promo_code": "Código promocional inválido",

		// --- Subscription ---
		"company_name_required": "El nombre de la empresa es obligatorio cuando es empresa",
		"no_active_plan":        "Sin plan activo",
		"max_users_reached":     "El tenant alcanzó el límite de usuarios (%d/%d)",
		"user_already_member":   "El usuario ya es miembro de este tenant",
		"role_not_found_tenant": "Rol '%s' no encontrado para este tenant",

		// --- Profile ---
		"profile_not_found":         "Perfil no encontrado",
		"profile_updated":           "Perfil actualizado",
		"failed_update_profile":     "Error al actualizar perfil",
		"tenant_profile_updated":    "Perfil del tenant actualizado",
		"only_owner_update_profile": "Solo el propietario puede actualizar el perfil del tenant",
		"only_owner_upload_logo":    "Solo el propietario puede subir el logo",

		// --- Upload ---
		"no_file_provided": "No se proporcionó archivo",
		"failed_upload":    "Error en la carga",

		// --- Members ---
		"failed_list_members":      "Error al listar miembros",
		"member_not_found":         "Miembro no encontrado",
		"member_added":             "Miembro agregado",
		"member_role_updated":      "Rol del miembro actualizado",
		"failed_update_role":       "Error al actualizar rol",
		"only_owner_remove":        "Solo el propietario puede eliminar miembros",
		"cannot_remove_self":       "No puede eliminarse a sí mismo",
		"member_removed":           "Miembro eliminado",
		"failed_remove_member":     "Error al eliminar miembro",
		"insufficient_permissions": "Permisos insuficientes",

		// --- Roles ---
		"failed_list_roles":       "Error al listar roles",
		"role_not_found":          "Rol no encontrado",
		"only_owner_manage_roles": "Solo el propietario puede gestionar roles",
		"failed_create_role":      "Error al crear rol",
		"role_updated":            "Rol actualizado",
		"role_deleted":            "Rol eliminado",

		// --- Permissions ---
		"failed_list_permissions":  "Error al listar permisos",
		"only_owner_manage_perms":  "Solo el propietario puede gestionar permisos",
		"permission_assigned":      "Permiso asignado",
		"permission_removed":       "Permiso eliminado",
		"failed_assign_permission": "Error al asignar permiso",
		"failed_remove_permission": "Error al eliminar permiso",

		// --- Features ---
		"feature_not_in_plan":    "El recurso '%s' no está disponible en su plan",
		"failed_list_features":   "Error al listar recursos",
		"feature_not_found":      "Recurso no encontrado",
		"feature_already_exists": "El recurso ya existe",
		"failed_create_feature":  "Error al crear recurso",
		"failed_update_feature":  "Error al actualizar recurso",
		"failed_delete_feature":  "Error al eliminar recurso",
		"failed_add_feature":     "Error al agregar recurso",
		"failed_remove_feature":  "Error al eliminar recurso",

		// --- Products ---
		"failed_list_products":  "Error al listar productos",
		"product_not_found":     "Producto no encontrado",
		"failed_create_product": "Error al crear producto",
		"product_updated":       "Producto actualizado",
		"failed_update_product": "Error al actualizar producto",
		"product_deleted":       "Producto eliminado",
		"failed_delete_product": "Error al eliminar producto",
		"failed_save_image":     "Error al guardar registro de imagen",

		// --- Services ---
		"failed_list_services":  "Error al listar servicios",
		"service_not_found":     "Servicio no encontrado",
		"failed_create_service": "Error al crear servicio",
		"service_updated":       "Servicio actualizado",
		"failed_update_service": "Error al actualizar servicio",
		"service_deleted":       "Servicio eliminado",
		"failed_delete_service": "Error al eliminar servicio",

		// --- Settings ---
		"layout_settings_saved": "Configuraciones de diseño guardadas",
		"settings_saved":        "Configuraciones guardadas",
		"failed_save_layout":    "Error al guardar configuraciones de diseño",
		"failed_save_settings":  "Error al guardar configuraciones",
		"language_updated":      "Idioma actualizado",
		"invalid_language":      "Idioma inválido. Use: pt-BR, pt, en, es",

		// --- App Users ---
		"failed_list_app_users":  "Error al listar app users",
		"app_user_not_found":     "App user no encontrado",
		"status_updated":         "Estado actualizado",
		"failed_update_status":   "Error al actualizar estado",
		"app_user_deleted":       "App user eliminado",
		"failed_delete_app_user": "Error al eliminar app user",

		// --- Admin specific ---
		"admin_not_found":          "Administrador no encontrado",
		"permission_denied":        "Permiso denegado",
		"email_already_exists":     "El correo electrónico ya existe",
		"failed_list_admins":       "Error al listar administradores",
		"failed_update":            "Error al actualizar",
		"failed_delete":            "Error al eliminar",
		"failed_update_password":   "Error al actualizar contraseña",
		"invalid_current_password": "Contraseña actual inválida",

		// --- Admin Roles ---
		"role_already_exists": "El rol ya existe",
		"failed_assign_role":  "Error al asignar rol",
		"failed_remove_role":  "Error al eliminar rol",
		"failed_delete_role":  "Error al eliminar rol",

		// --- Admin Tenants ---
		"failed_list_tenants":        "Error al listar tenants",
		"tenant_already_exists":      "El tenant ya existe",
		"plan_not_found":             "Plan no encontrado",
		"transaction_failed":         "Error en la transacción",
		"failed_create_plan":         "Error al crear plan",
		"owner_creation_failed":      "Error al crear propietario: %s",
		"commit_failed":              "Error al confirmar transacción",
		"failed_update_tenant":       "Error al actualizar tenant",
		"failed_delete_tenant":       "Error al eliminar tenant",
		"failed_update_status_admin": "Error al actualizar estado",
		"failed_get_history":         "Error al obtener historial",
		"failed_list_members_admin":  "Error al listar miembros",

		// --- Admin Plans ---
		"failed_list_plans_admin":  "Error al listar planes",
		"failed_create_plan_admin": "Error al crear plan",
		"plan_not_found_admin":     "Plan no encontrado",
		"failed_update_plan":       "Error al actualizar plan",
		"failed_delete_plan":       "Error al eliminar plan",

		// --- Admin Promotions ---
		"failed_list_promotions":      "Error al listar promociones",
		"failed_create_promotion":     "Error al crear promoción",
		"promotion_not_found":         "Promoción no encontrada",
		"failed_update_promotion":     "Error al actualizar promoción",
		"failed_deactivate_promotion": "Error al desactivar promoción",

		// --- Validation templates ---
		"validation.required": "%s es obligatorio",
		"validation.email":    "%s debe ser un correo electrónico válido",
		"validation.min":      "%s debe tener al menos %s caracteres",
		"validation.max":      "%s debe tener como máximo %s caracteres",
		"validation.url":      "%s debe ser una URL válida",
		"validation.oneof":    "%s debe ser uno de: %s",
		"validation.uuid":     "%s debe ser un UUID válido",
		"validation.gte":      "%s debe ser mayor o igual a %s",
		"validation.lte":      "%s debe ser menor o igual a %s",
		"validation.len":      "%s debe tener exactamente %s caracteres",
		"validation.default":  "%s es inválido",
	},
}

// fieldLabels contains localized field labels for validation messages.
var fieldLabels = map[string]map[string]string{
	LangPtBR: {
		"name": "Nome", "email": "E-mail", "password": "Senha",
		"plan_id": "Plano", "billing_cycle": "Ciclo de cobrança", "subdomain": "Subdomínio",
		"company_name": "Nome da empresa", "is_company": "É empresa", "promo_code": "Código promocional",
		"title": "Título", "slug": "Slug", "role_id": "Função", "permission_id": "Permissão",
		"category": "Categoria", "status": "Status", "current_password": "Senha atual",
		"new_password": "Nova senha", "token": "Token", "tenant_id": "Tenant",
		"full_name": "Nome completo", "description": "Descrição", "price": "Preço",
		"sku": "SKU", "stock": "Estoque", "duration": "Duração", "language": "Idioma",
	},
	LangPt: {
		"name": "Nome", "email": "E-mail", "password": "Palavra-passe",
		"plan_id": "Plano", "billing_cycle": "Ciclo de cobrança", "subdomain": "Subdomínio",
		"company_name": "Nome da empresa", "is_company": "É empresa", "promo_code": "Código promocional",
		"title": "Título", "slug": "Slug", "role_id": "Função", "permission_id": "Permissão",
		"category": "Categoria", "status": "Estado", "current_password": "Palavra-passe atual",
		"new_password": "Nova palavra-passe", "token": "Token", "tenant_id": "Tenant",
		"full_name": "Nome completo", "description": "Descrição", "price": "Preço",
		"sku": "SKU", "stock": "Stock", "duration": "Duração", "language": "Idioma",
	},
	LangEn: {
		"name": "Name", "email": "Email", "password": "Password",
		"plan_id": "Plan", "billing_cycle": "Billing cycle", "subdomain": "Subdomain",
		"company_name": "Company name", "is_company": "Is company", "promo_code": "Promo code",
		"title": "Title", "slug": "Slug", "role_id": "Role", "permission_id": "Permission",
		"category": "Category", "status": "Status", "current_password": "Current password",
		"new_password": "New password", "token": "Token", "tenant_id": "Tenant",
		"full_name": "Full name", "description": "Description", "price": "Price",
		"sku": "SKU", "stock": "Stock", "duration": "Duration", "language": "Language",
	},
	LangEs: {
		"name": "Nombre", "email": "Correo electrónico", "password": "Contraseña",
		"plan_id": "Plan", "billing_cycle": "Ciclo de facturación", "subdomain": "Subdominio",
		"company_name": "Nombre de la empresa", "is_company": "Es empresa", "promo_code": "Código promocional",
		"title": "Título", "slug": "Slug", "role_id": "Rol", "permission_id": "Permiso",
		"category": "Categoría", "status": "Estado", "current_password": "Contraseña actual",
		"new_password": "Nueva contraseña", "token": "Token", "tenant_id": "Tenant",
		"full_name": "Nombre completo", "description": "Descripción", "price": "Precio",
		"sku": "SKU", "stock": "Stock", "duration": "Duración", "language": "Idioma",
	},
}
