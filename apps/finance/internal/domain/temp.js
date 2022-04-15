const constants = {
  DEV_STRIPE_CUSTOMER_ID: 'dev-customer-id',
  ROBOT_STRIPE_INTENT_ID: 'robot-intent-id',
  ROBOT_STRIPE_CUSTOMER_ID: 'robot-customer-id',
};

const robotBilling = {
  stripeCustomerId: constants.ROBOT_STRIPE_CUSTOMER_ID,
  stripeSetupIntentId: constants.ROBOT_STRIPE_INTENT_ID,
  cardholderName: 'Robot',
  address: {
    line1: 'unknown',
  },
};

export const makeAccountDomain = ({
  accountRepo = dbApi(),
  iamSvc = makeIAMSvcAdapter(),
  stripeSvc = makeStripeSvcAdapter(),
  authSvc = makeAuthSvcAdapter(),
}) => {
  const ensureAccount = async (accountId) => {
    assert(accountId, ErrBadParams('accountId is required'));
    const account = await accountRepo.findOne({ id: accountId });
    assert(
      account?.isActive,
      ErrBadRequest('Account not found or is deactivated')
    );
    return account;
  };

  const accountSvc = {};

  accountSvc.listAccounts = async (ctx) => {
    const myMemberships = await iamSvc.myMemberships(ctx, {
      resourceType: IAM_ACCOUNT_RESOURCE_TYPE,
    });

    console.log('my memberships', myMemberships, ctx.session);

    return accountRepo.find({
      id: { $in: myMemberships.map((rb) => rb.resourceId) },
    });
  };

  accountSvc.getAccount = async (
    ctx,
    { accountId, fromChaining = false } = {}
  ) => {
    assert(accountId, ErrBadParams('accountId is required'));
    const canI = await iamSvc.canI(ctx, {
      resourceIds: [accountId],
      action: ACTION_ACCOUNT_GET,
    });

    if (!fromChaining) assert(canI, 'You are not allowed to get this account');
    return ensureAccount(accountId);
  };

  accountSvc.getStripeSetupIntent = async () => {
    return stripeSvc.getSetupIntent();
  };

  accountSvc.createAccount = async (
    ctx,
    { name, billing = null, isRobot } = {}
  ) => {
    assert(name, ErrBadParams('name is required'));
    if (!isRobot) assert(billing, ErrBadParams('billing is required'));

    // TODO: think of a better way to do this
    const accountId = accountRepo.newId();

    console.log('ctx.session: ', ctx.session);

    const customerId = await (async () => {
      try {
        if (isRobot) return null;
        return stripeSvc.createCustomer({
          accountId,
          paymentMethod: billing.stripePaymentMethod,
        });
      } catch (err) {
        throw createHttpError(
          StatusCodes.BAD_REQUEST,
          `could not create stripe customer because ${err.message}`
        );
      }
    })();

    const account = await (async () => {
      try {
        const result = await accountRepo.create({
          id: accountId,
          name,
          contactEmail: ctx.session?.userEmail,
          isRobot,
          billing: isRobot
            ? robotBilling
            : { ...(billing || {}), stripeCustomerId: customerId },
        });

        return result;
      } catch (err) {
        throw createHttpError(
          StatusCodes.BAD_REQUEST,
          `could not create account because ${err.message}`
        );
      }
    })();

    await iamSvc.addMembership(ctx, {
      userId: ctx.session?.userId,
      resourceType: IAM_ACCOUNT_RESOURCE_TYPE,
      resourceId: account.id,
      role: ROLE_ACCOUNT_OWNER,
    });

    return account;
  };

  accountSvc.listMemberships = async (
    ctx,
    { accountId } = {},
    { skipAccountData = false } = {}
  ) => {
    await ensureAccount(accountId);

    const memberships = await iamSvc.resourceMemberships(ctx, {
      resourceId: accountId,
    });

    if (skipAccountData) return memberships;

    const accounts = await accountRepo.find({
      id: { $in: memberships.map((m) => m.resourceId) },
    });

    const accountsMap = accounts.reduce(
      (acc, curr) => ({ ...acc, [curr.id]: curr }),
      {}
    );

    return memberships.map((m) => {
      return {
        ...m,
        account: accountsMap[m.resourceId],
      };
    });
  };

  accountSvc.updateAccount = async (ctx, { accountId, name, contactEmail }) => {
    await ensureAccount(accountId);

    const canI = await iamSvc.canI(ctx, {
      resourceIds: [accountId],
      action: ACTION_ACCOUNT_UPDATE,
    });

    assert(
      canI,
      ErrUnAuthorized('you are not authorized to update this account')
    );

    // TODO: assert if contactEmail belongs to one of its members

    const resourceMemberships = await iamSvc.resourceMemberships(ctx, {
      resourceId: accountId,
    });

    const ownerMembers = resourceMemberships.filter(
      (item) => item.role === ROLE_ACCOUNT_OWNER
    );

    const userByEmail = await authSvc.findByEmail(ctx, { email: contactEmail });
    assert(userByEmail?.id, 'user with contactEmail does not exist');
    assert(
      ownerMembers.map((item) => item.userId).includes(userByEmail?.id),
      'contactEmail is not an owner of this account'
    );

    return accountRepo.update(
      { id: accountId },
      {
        name,
        contactEmail,
      }
    );
  };

  accountSvc.updateBilling = async (
    ctx,
    { accountId, billing, skipStripe = false }
  ) => {
    assert(accountId, ErrBadParams('accountId is required'));
    assert(billing, ErrBadParams('billing is required'));

    const account = await ensureAccount(accountId);

    const customerId = (() => {
      if (skipStripe) return constants.DEV_STRIPE_CUSTOMER_ID;
      return stripeSvc.createCustomer({
        accountId,
        paymentMethod: billing.stripePaymentMethod,
      });
    })();

    assert(customerId, ErrBadParams('could not create billing customer'));

    if (account.billing?.stripeCustomerId !== constants.DEV_STRIPE_CUSTOMER_ID)
      await stripeSvc.deleteCustomer({
        customerId: account.billing.stripeCustomerId,
      });

    return accountRepo.update(
      { id: accountId },
      {
        billing: {
          ...billing,
          stripeCustomerId: customerId,
        },
      }
    );
  };

  accountSvc.deactivate = async (ctx, { accountId } = {}) => {
    assert(accountId, ErrBadParams('accountId is required'));
    await ensureAccount(accountId);
    const acc = await accountRepo.update({ id: accountId }, { active: false });
    return acc.active === false;
  };

  accountSvc.activate = async (ctx, { accountId } = {}) => {
    assert(accountId, ErrBadParams('accountId is required'));
    const acc = await accountRepo.update({ id: accountId }, { active: true });
    return acc.active === true;
  };

  accountSvc.inviteMember = async (
    ctx,
    { accountId, name, email, role } = {}
  ) => {
    assert(accountId, 'accountId is required');
    assert(name, 'name is required');
    assert(email, 'email is required');
    assert(role, 'role is required');

    await ensureAccount(accountId);

    const canInvite = await iamSvc.canI(ctx, {
      action: ACTION_ACCOUNT_INVITE_MEMBER,
      resourceIds: [accountId],
    });

    assert(canInvite, 'You are not allowed to invite members to this account');

    const invitedUserId = await authSvc.inviteSignup(ctx, { email, name });
    await iamSvc.addMembership(ctx, {
      userId: invitedUserId,
      resourceType: IAM_ACCOUNT_RESOURCE_TYPE,
      resourceId: accountId,
      role,
    });

    return true;
  };

  accountSvc.updateMember = async (ctx, { accountId, userId, role }) => {
    assert(accountId, ErrBadParams('accountId is required'));
    assert(userId, ErrBadParams('userId is required'));
    assert(role, ErrBadParams('role is required'));

    const account = await ensureAccount(accountId);

    const canUpdate = await iamSvc.canI(ctx, {
      resourceIds: [accountId],
      action: ACTION_ACCOUNT_UPDATE_MEMBER,
    });

    assert(canUpdate, 'You are not allowed to update members of this account');

    const accContactUser = await authSvc.findByEmail(ctx, {
      email: account.contactEmail,
    });

    assert(
      accContactUser?.id !== userId,
      ErrBadParams("can't update role of member who is the account contact")
    );

    await iamSvc.removeMembership(ctx, { userId, resourceId: accountId });
    await iamSvc.addMembership(ctx, {
      userId,
      resourceType: IAM_ACCOUNT_RESOURCE_TYPE,
      resourceId: accountId,
      role,
    });

    return true;
  };

  accountSvc.removeMember = async (ctx, { accountId, userId }) => {
    assert(accountId, ErrBadParams('accountId is required'));
    assert(userId, ErrBadParams('userId is required'));

    const account = await ensureAccount(accountId);

    const canRemove = await iamSvc.canI(ctx, {
      action:
        userId === ctx.session.userId
          ? ACTION_ACCOUNT_REMOVE_MEMBER_SELF
          : ACTION_ACCOUNT_REMOVE_MEMBER,
      resourceIds: [accountId],
    });

    assert(
      canRemove,
      ErrUnAuthorized('You are not allowed to remove members of this account')
    );

    const accContactUser = await authSvc.findByEmail(ctx, {
      email: account.contactEmail,
    });

    assert(
      accContactUser?.id !== userId,
      ErrUnAuthorized("you can't remove member who is the account contact")
    );

    await iamSvc.removeMembership(ctx, {
      resourceId: accountId,
      userId,
    });
    return true;
  };

  accountSvc.deleteAccount = async (ctx, { accountId } = {}) => {
    assert(accountId, ErrBadParams('accountId is required'));

    const canI = await iamSvc.canI(ctx, {
      resourceIds: [accountId],
      action: ACTION_ACCOUNT_DELETE,
    });

    assert(canI, ErrUnAuthorized('You are not allowed to delete this account'));
    await iamSvc.removeResource(ctx, { resourceId: accountId });
    await accountRepo.delete({ id: accountId });
    return true;
  };

  return accountSvc;
};
