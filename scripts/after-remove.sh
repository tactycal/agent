# Read defaults, if available:
[ -f "/etc/default/tactycal-agent" ] && . /etc/default/tactycal-agent

# Sane defaults:
[ -z "$AGENT_HOME" ]       && AGENT_HOME=/var/spool/carbon-c-relay
[ -z "$AGENT_USER" ]       && AGENT_USER=tactycal
[ -z "$AGENT_GROUP" ]      && AGENT_GROUP=tactycal
[ -z "$AGENT_CONF_DIR_PATH" ]  && AGENT_CONF_DIR_PATH=/etc/tactycal
[ -z "$AGENT_CONF_PATH" ]      && AGENT_CONF_PATH=$AGENT_CONF_DIR_PATH/agent.conf
[ -z "$AGENT_STATE_DIR_PATH" ] && AGENT_STATE_DIR_PATH=/var/opt/tactycal
[ -z "$OWNER_USER" ]       && OWNER_USER=root

# 1. Remove configuration folder
test -d $AGENT_CONF_DIR_PATH && rm -fr $AGENT_CONF_DIR_PATH

# 2. Remove configuration file
test -g $AGENT_CONF_PATH && rm -f $AGENT_CONF_PATH

# 3. Remove state folder
test -d $AGENT_STATE_DIR_PATH && rm -fr $AGENT_STATE_DIR_PATH

# 4. Remove user
if ! getent passwd | grep -q "^$AGENT_USER:"; then
    echo -n "Removing system user $AGENT_USER.."
    userdel $AGENT_USER >/dev/null 2>/dev/null || true
    echo "..done"
fi

# 5. Remove group
if getent group | grep -q "^$AGENT_GROUP:" ; then
    echo -n "Removing group $AGENT_GROUP.."
    groupdel $AGENT_GROUP >/dev/null 2>/dev/null || true
    echo "..done"
fi
