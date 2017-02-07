# Read defaults, if available:
[ -f "/etc/default/tactycal-agent" ] && . /etc/default/tactycal-agent

# Sane defaults:
[ -z "$AGENT_HOME" ]           && AGENT_HOME=/var/spool/carbon-c-relay
[ -z "$AGENT_USER" ]           && AGENT_USER=tactycal
[ -z "$AGENT_NAME" ]           && AGENT_NAME="tactycal agent"
[ -z "$AGENT_GROUP" ]          && AGENT_GROUP=tactycal
[ -z "$AGENT_CONF_DIR_PATH" ]  && AGENT_CONF_DIR_PATH=/etc/tactycal
[ -z "$AGENT_CONF_PATH" ]      && AGENT_CONF_PATH=$AGENT_CONF_DIR_PATH/agent.conf
[ -z "$AGENT_STATE_DIR_PATH" ] && AGENT_STATE_DIR_PATH=/var/opt/tactycal
[ -z "$OWNER_USER" ]           && OWNER_USER=root

# create user to avoid running agent as root
# 1. create group if not existing
if ! getent group | grep -q "^$AGENT_GROUP:" ; then
   echo -n "Adding group $AGENT_GROUP.."
   groupadd --system $AGENT_GROUP >/dev/null 2>/dev/null || true
   echo "..done"
fi

# 2. create user if not existing
if ! getent passwd | grep -q "^$AGENT_USER:"; then
  echo -n "Adding system user $AGENT_USER.."
  useradd --system \
          --gid $AGENT_GROUP \
          --no-create-home \
          $AGENT_USER >/dev/null 2>/dev/null || true
  echo "..done"
fi

# 3. create default configuration folder
test -d $AGENT_CONF_DIR_PATH || mkdir $AGENT_CONF_DIR_PATH

# 4. create default state folder
test -d $AGENT_STATE_DIR_PATH || mkdir -p $AGENT_STATE_DIR_PATH

# 5. Add empty configuration file
test -f $AGENT_CONF_PATH || touch $AGENT_CONF_PATH

# 6. adjust file and directory permissions
chown -R $OWNER_USER:$AGENT_GROUP $AGENT_CONF_DIR_PATH
chmod u=rwx,g=rx,o= $AGENT_CONF_DIR_PATH

chown -R $OWNER_USER:$AGENT_GROUP $AGENT_CONF_PATH
chmod u=rw,g=r,o= $AGENT_CONF_PATH

chown -R $OWNER_USER:$AGENT_GROUP $AGENT_STATE_DIR_PATH
chmod u=rwx,g=rwx,o= $AGENT_STATE_DIR_PATH
