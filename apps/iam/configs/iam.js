const iam = {
  account: {
    roles: ['account-owner', 'account-admin', 'account-member'],
    actions: {
      // account actions
      invite_acc_member: ['account-owner', 'account-admin'],
      invite_acc_admin: ['account-owner'],
      update_payment: ['account-owner'],
      create_project: ['account-admin', 'account-owner'],
      list_projects: ['account-admin', 'account-owner', 'account-member'],
      delete_payment: ['account-owner'],
      add_domain: ['account-owner', 'account-admin'],
      update_domain: ['account-owner', 'account-admin'],
      delete_domain: ['account-owner', 'account-admin'],
      list_domain: ['account-owner', 'account-admin', 'account-member'],
      check_invoices: ['account-owner'],
      check_outstanding: ['account-owner'],
      update_account: ['account-owner'],
      delete_account: ['account-owner'],
      update_acc_member: ['account-owner', 'account-admin'],
      update_acc_admin: ['account-owner'],
      cancel_acc_invite: ['account-owner', 'account-admin'],
      list_devices: ['account-owner', 'account-admin'],

      // projects actions
      read_project: ['account-owner', 'account-admin', 'account-member'],
      update_project: ['account-owner', 'account-admin', 'account-member'],
      delete_project: ['account-admin', 'account-owner'],
      invite_proj_admin: ['account-owner', 'account-admin'],
      invite_proj_member: ['account-owner', 'account-admin', 'account-member'],
      invite_proj_owner: ['account-owner'],
      cancel_proj_invite: ['account-owner', 'account-admin'],
      get_docker_credentials: ['account-owner', 'account-admin', 'account-member'],

      // device actions
      update_device: [],
      delete_device: ['account-owner', 'account-admin'],
    },
  },
  project: {
    roles: ['project-owner', 'project-admin', 'project-member'],
    actions: {
      read_project: ['project-owner', 'project-admin', 'project-member'],
      update_project: ['project-owner', 'project-admin', 'project-member'],
      delete_project: ['project-owner', 'project-admin'],
      invite_proj_admin: ['project-owner', 'project-admin'],
      invite_proj_member: ['project-owner', 'project-admin', 'project-member'],
      invite_proj_owner: ['project-owner'],
      cancel_proj_invite: ['project-owner', 'project-admin'],
      get_docker_credentials: ['project-owner', 'project-admin','project-member'],

      // device actions
      update_device: [],
      delete_device: [],
    },
  },
  device: {
    roles: ['device-owner', 'account-owner', 'account-admin'],
    actions: {
      update_device: ['device-owner'],
      delete_device: ['device-owner'],
    },
  },
};

const main = () => {
  const iamConfig = Object.values(iam).reduce(
    ({ actions }, { actions: a }) => {
      return {
        actions: {
          ...actions,
          ...Object.entries(a).reduce((ac, [k, v]) => {
            return {
              ...ac,
              [k]: [...(actions[k] || []), ...v],
            };
          }, {}),
        },
      };
    },
    { actions: {} }
  );
  console.log(JSON.stringify(iamConfig, null, 2));
};

main();
